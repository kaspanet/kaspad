package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blocks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	mathUtil "github.com/kaspanet/kaspad/util/math"
	"github.com/pkg/errors"
)

// RelayInvsContext is the interface for the context needed for the HandleRelayInvs flow.
type RelayInvsContext interface {
	Domain() domain.Domain
	NetAdapter() *netadapter.NetAdapter
	OnNewBlock(block *externalapi.DomainBlock) error
	SharedRequestedBlocks() *SharedRequestedBlocks
	StartIBDIfRequired() error
	IsInIBD() bool
	Broadcast(message appmessage.Message) error
	AddOrphan(orphanBlock *externalapi.DomainBlock)
	IsOrphan(blockHash *externalapi.DomainHash) bool
}

type handleRelayInvsFlow struct {
	RelayInvsContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
	invsQueue                    []*appmessage.MsgInvRelayBlock
}

// HandleRelayInvs listens to appmessage.MsgInvRelayBlock messages, requests their corresponding blocks if they
// are missing, adds them to the DAG and propagates them to the rest of the network.
func HandleRelayInvs(context RelayInvsContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	flow := &handleRelayInvsFlow{
		RelayInvsContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
		peer:             peer,
		invsQueue:        make([]*appmessage.MsgInvRelayBlock, 0),
	}
	return flow.start()
}

func (flow *handleRelayInvsFlow) start() error {
	for {
		log.Debugf("Waiting for inv")
		inv, err := flow.readInv()
		if err != nil {
			return err
		}

		log.Debugf("Got relay inv for block %s", inv.Hash)

		blockInfo, err := flow.Domain().Consensus().GetBlockInfo(inv.Hash)
		if err != nil {
			return err
		}
		if blockInfo.Exists {
			if blockInfo.BlockStatus == externalapi.StatusInvalid {
				return protocolerrors.Errorf(true, "sent inv of an invalid block %s",
					inv.Hash)
			}
			log.Debugf("Block %s already exists. continuing...", inv.Hash)
			continue
		}

		if flow.IsOrphan(inv.Hash) {
			log.Debugf("Block %s is a known orphan. continuing...", inv.Hash)
			continue
		}

		err = flow.StartIBDIfRequired()
		if err != nil {
			return err
		}
		if flow.IsInIBD() {
			log.Debugf("Got block %s while in IBD. continuing...", inv.Hash)
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

func (flow *handleRelayInvsFlow) readInv() (*appmessage.MsgInvRelayBlock, error) {
	if len(flow.invsQueue) > 0 {
		var inv *appmessage.MsgInvRelayBlock
		inv, flow.invsQueue = flow.invsQueue[0], flow.invsQueue[1:]
		return inv, nil
	}

	msg, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}

	inv, ok := msg.(*appmessage.MsgInvRelayBlock)
	if !ok {
		return nil, protocolerrors.Errorf(true, "unexpected %s message in the block relay handleRelayInvsFlow while "+
			"expecting an inv message", msg.Command())
	}
	return inv, nil
}

func (flow *handleRelayInvsFlow) requestBlocks(requestQueue *hashesQueueSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "handleRelayInvsFlow.requestBlocks")
	defer onEnd()

	numHashesToRequest := mathUtil.MinInt(appmessage.MaxRequestRelayBlocksHashes, requestQueue.len())
	hashesToRequest := requestQueue.dequeue(numHashesToRequest)

	requestedBlocksSet := map[externalapi.DomainHash]struct{}{}
	var filteredHashesToRequest []*externalapi.DomainHash
	for _, hash := range hashesToRequest {
		exists := flow.SharedRequestedBlocks().addIfNotExists(hash)
		if exists {
			continue
		}

		// The block can become known from another peer in the process of orphan resolution
		blockInfo, err := flow.Domain().Consensus().GetBlockInfo(hash)
		if err != nil {
			return err
		}
		if blockInfo.Exists {
			continue
		}

		requestedBlocksSet[*hash] = struct{}{}
		filteredHashesToRequest = append(filteredHashesToRequest, hash)
	}

	// Exit early if we've filtered out all the hashes
	if len(filteredHashesToRequest) == 0 {
		log.Debugf("requestBlocks - no hashes to request. Returning..")
		return nil
	}

	// In case the function returns earlier than expected, we want to make sure flow.SharedRequestedBlocks() is
	// clean from any pending blocks.
	defer flow.SharedRequestedBlocks().removeSet(requestedBlocksSet)

	getRelayBlocksMsg := appmessage.NewMsgRequestRelayBlocks(filteredHashesToRequest)
	err := flow.outgoingRoute.Enqueue(getRelayBlocksMsg)
	if err != nil {
		return err
	}

	for _, expectedHash := range filteredHashesToRequest {
		log.Debugf("Waiting for block %s", expectedHash)
		msgBlock, err := flow.readMsgBlock()
		if err != nil {
			return err
		}

		block := appmessage.MsgBlockToDomainBlock(msgBlock)
		blockHash := consensusserialization.BlockHash(block)
		log.Criticalf("got block %s", blockHash)

		if *blockHash != *expectedHash {
			return protocolerrors.Errorf(true, "expected block %s but got %s", expectedHash, blockHash)
		}

		err = flow.processAndRelayBlock(requestQueue, requestedBlocksSet, block)
		if err != nil {
			return err
		}

		flow.SharedRequestedBlocks().remove(blockHash)
	}
	return nil
}

// readMsgBlock returns the next msgBlock in msgChan, and populates invsQueue with any inv messages that meanwhile arrive.
//
// Note: this function assumes msgChan can contain only appmessage.MsgInvRelayBlock and appmessage.MsgBlock messages.
func (flow *handleRelayInvsFlow) readMsgBlock() (msgBlock *appmessage.MsgBlock, err error) {
	for {
		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return nil, err
		}

		switch message := message.(type) {
		case *appmessage.MsgInvRelayBlock:
			flow.invsQueue = append(flow.invsQueue, message)
		case *appmessage.MsgBlock:
			return message, nil
		default:
			return nil, errors.Errorf("unexpected message %s", message.Command())
		}
	}
}

func (flow *handleRelayInvsFlow) processAndRelayBlock(requestQueue *hashesQueueSet, requestedBlocksSet map[externalapi.DomainHash]struct{}, block *externalapi.DomainBlock) error {
	blockHash := consensusserialization.BlockHash(block)
	log.Debugf("processAndRelayBlock start for block %s", blockHash)
	defer log.Debugf("processAndRelayBlock end for block %s", blockHash)

	err := flow.Domain().Consensus().ValidateAndInsertBlock(block)
	if err != nil {
		if !errors.As(err, &ruleerrors.RuleError{}) {
			return errors.Wrapf(err, "failed to process block %s", blockHash)
		}

		missingParentsError := &ruleerrors.ErrMissingParents{}
		if errors.As(err, missingParentsError) {
			blueScore, err := blocks.ExtractBlueScore(block)
			if err != nil {
				return protocolerrors.Errorf(true, "received an orphan "+
					"block %s with malformed blue score", blockHash)
			}

			const maxOrphanBlueScoreDiff = 10000
			virtualSelectedParent, err := flow.Domain().Consensus().GetVirtualSelectedParent()
			if err != nil {
				return err
			}
			selectedTipBlueScore, err := blocks.ExtractBlueScore(virtualSelectedParent)
			if err != nil {
				return err
			}

			if blueScore > selectedTipBlueScore+maxOrphanBlueScoreDiff {
				log.Infof("Orphan block %s has blue score %d and the selected tip blue score is "+
					"%d. Ignoring orphans with a blue score difference from the selected tip greater than %d",
					blockHash, blueScore, selectedTipBlueScore, maxOrphanBlueScoreDiff)
				return nil
			}

			// Add the orphan to the orphan pool
			flow.AddOrphan(block)

			// Request the parents for the orphan block from the peer that sent it.
			for _, missingAncestor := range missingParentsError.MissingParentHashes {
				if _, ok := requestedBlocksSet[*missingAncestor]; ok {
					log.Debugf("ancestor %s already exists in requestedBlocksSet", missingAncestor)
					continue
				}
				requestQueue.enqueueIfNotExists(missingAncestor)
			}
			return nil
		}
		log.Infof("Rejected block %s from %s: %s", blockHash, flow.peer, err)

		return protocolerrors.Wrapf(true, err, "got invalid block %s from relay", blockHash)
	}

	err = flow.Broadcast(appmessage.NewMsgInvBlock(blockHash))
	if err != nil {
		return err
	}

	log.Debugf("Accepted block %s via relay", blockHash)

	err = flow.StartIBDIfRequired()
	if err != nil {
		return err
	}
	err = flow.OnNewBlock(block)
	if err != nil {
		return err
	}

	return nil
}
