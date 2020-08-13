package blockrelay

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/blocklogger"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	mathUtil "github.com/kaspanet/kaspad/util/math"
	"github.com/pkg/errors"
)

// RelayInvsContext is the interface for the context needed for the HandleRelayInvs flow.
type RelayInvsContext interface {
	NetAdapter() *netadapter.NetAdapter
	DAG() *blockdag.BlockDAG
	OnNewBlock(block *util.Block) error
	SharedRequestedBlocks() *SharedRequestedBlocks
	StartIBDIfRequired()
	IsInIBD() bool
	Broadcast(message domainmessage.Message) error
}

type handleRelayInvsFlow struct {
	RelayInvsContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
	invsQueue                    []*domainmessage.MsgInvRelayBlock
}

// HandleRelayInvs listens to domainmessage.MsgInvRelayBlock messages, requests their corresponding blocks if they
// are missing, adds them to the DAG and propagates them to the rest of the network.
func HandleRelayInvs(context RelayInvsContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	flow := &handleRelayInvsFlow{
		RelayInvsContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
		peer:             peer,
		invsQueue:        make([]*domainmessage.MsgInvRelayBlock, 0),
	}
	return flow.start()
}

func (flow *handleRelayInvsFlow) start() error {
	for {
		inv, err := flow.readInv()
		if err != nil {
			return err
		}

		if flow.DAG().IsKnownBlock(inv.Hash) {
			if flow.DAG().IsKnownInvalid(inv.Hash) {
				return protocolerrors.Errorf(true, "sent inv of an invalid block %s",
					inv.Hash)
			}
			continue
		}

		flow.StartIBDIfRequired()
		if flow.IsInIBD() {
			// Block relay is disabled during IBD
			continue
		}

		requestQueue := newHashesQueueSet()
		requestQueue.enqueueIfNotExists(inv.Hash)

		for requestQueue.len() > 0 {
			err := flow.requestBlocks(requestQueue)
			if err != nil {
				return err
			}
		}
	}
}

func (flow *handleRelayInvsFlow) readInv() (*domainmessage.MsgInvRelayBlock, error) {

	if len(flow.invsQueue) > 0 {
		var inv *domainmessage.MsgInvRelayBlock
		inv, flow.invsQueue = flow.invsQueue[0], flow.invsQueue[1:]
		return inv, nil
	}

	msg, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}

	inv, ok := msg.(*domainmessage.MsgInvRelayBlock)
	if !ok {
		return nil, protocolerrors.Errorf(true, "unexpected %s message in the block relay handleRelayInvsFlow while "+
			"expecting an inv message", msg.Command())
	}
	return inv, nil
}

func (flow *handleRelayInvsFlow) requestBlocks(requestQueue *hashesQueueSet) error {
	numHashesToRequest := mathUtil.MinInt(domainmessage.MsgRequestRelayBlocksHashes, requestQueue.len())
	hashesToRequest := requestQueue.dequeue(numHashesToRequest)

	pendingBlocks := map[daghash.Hash]struct{}{}
	var filteredHashesToRequest []*daghash.Hash
	for _, hash := range hashesToRequest {
		exists := flow.SharedRequestedBlocks().addIfNotExists(hash)
		if exists {
			continue
		}

		pendingBlocks[*hash] = struct{}{}
		filteredHashesToRequest = append(filteredHashesToRequest, hash)
	}

	// Exit early if we've filtered out all the hashes
	if len(filteredHashesToRequest) == 0 {
		return nil
	}

	// In case the function returns earlier than expected, we want to make sure requestedBlocks is
	// clean from any pending blocks.
	defer flow.SharedRequestedBlocks().removeSet(pendingBlocks)

	getRelayBlocksMsg := domainmessage.NewMsgRequestRelayBlocks(filteredHashesToRequest)
	err := flow.outgoingRoute.Enqueue(getRelayBlocksMsg)
	if err != nil {
		return err
	}

	for len(pendingBlocks) > 0 {
		msgBlock, err := flow.readMsgBlock()
		if err != nil {
			return err
		}

		block := util.NewBlock(msgBlock)
		blockHash := block.Hash()

		if _, ok := pendingBlocks[*blockHash]; !ok {
			return protocolerrors.Errorf(true, "got unrequested block %s", block.Hash())
		}

		err = flow.processAndRelayBlock(requestQueue, block)
		if err != nil {
			return err
		}

		delete(pendingBlocks, *blockHash)
		flow.SharedRequestedBlocks().remove(blockHash)

	}
	return nil
}

// readMsgBlock returns the next msgBlock in msgChan, and populates invsQueue with any inv messages that meanwhile arrive.
//
// Note: this function assumes msgChan can contain only domainmessage.MsgInvRelayBlock and domainmessage.MsgBlock messages.
func (flow *handleRelayInvsFlow) readMsgBlock() (
	msgBlock *domainmessage.MsgBlock, err error) {

	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}

		switch message := message.(type) {
		case *domainmessage.MsgInvRelayBlock:
			flow.invsQueue = append(flow.invsQueue, message)
		case *domainmessage.MsgBlock:
			return message, nil
		default:
			return nil, errors.Errorf("unexpected message %s", message.Command())
		}
	}
}

func (flow *handleRelayInvsFlow) processAndRelayBlock(requestQueue *hashesQueueSet, block *util.Block) error {
	blockHash := block.Hash()
	isOrphan, isDelayed, err := flow.DAG().ProcessBlock(block, blockdag.BFNone)
	if err != nil {
		if !errors.As(err, &blockdag.RuleError{}) {
			return errors.Wrapf(err, "failed to process block %s", blockHash)
		}
		log.Infof("Rejected block %s from %s: %s", blockHash, flow.peer, err)

		return protocolerrors.Wrapf(true, err, "got invalid block %s", blockHash)
	}

	if isDelayed {
		return nil
	}

	if isOrphan {
		blueScore, err := block.BlueScore()
		if err != nil {
			return protocolerrors.Errorf(true, "received an orphan "+
				"block %s with malformed blue score", blockHash)
		}

		const maxOrphanBlueScoreDiff = 10000
		selectedTipBlueScore := flow.DAG().SelectedTipBlueScore()
		if blueScore > selectedTipBlueScore+maxOrphanBlueScoreDiff {
			log.Infof("Orphan block %s has blue score %d and the selected tip blue score is "+
				"%d. Ignoring orphans with a blue score difference from the selected tip greater than %d",
				blockHash, blueScore, selectedTipBlueScore, maxOrphanBlueScoreDiff)
			return nil
		}

		// Request the parents for the orphan block from the peer that sent it.
		missingAncestors := flow.DAG().GetOrphanMissingAncestorHashes(blockHash)
		for _, missingAncestor := range missingAncestors {
			requestQueue.enqueueIfNotExists(missingAncestor)
		}
		return nil
	}
	err = blocklogger.LogBlock(block)
	if err != nil {
		return err
	}
	err = flow.Broadcast(domainmessage.NewMsgInvBlock(blockHash))
	if err != nil {
		return err
	}

	flow.StartIBDIfRequired()
	err = flow.OnNewBlock(block)
	if err != nil {
		return err
	}

	return nil
}
