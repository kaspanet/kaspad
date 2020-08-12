package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// HandleIBDContext is the interface for the context needed for the HandleIBD flow.
type HandleIBDContext interface {
	DAG() *blockdag.BlockDAG
	OnNewBlock(block *util.Block) error
	StartIBDIfRequired()
	FinishIBD()
}

type handleIBDFlow struct {
	HandleIBDContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// HandleIBD waits for IBD start and handles it when IBD is triggered for this peer
func HandleIBD(context HandleIBDContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

	flow := &handleIBDFlow{
		HandleIBDContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
		peer:             peer,
	}
	return flow.start()
}

func (flow *handleIBDFlow) start() error {
	for {
		err := flow.runIBD()
		if err != nil {
			return err
		}
	}
}

func (flow *handleIBDFlow) runIBD() error {
	flow.peer.WaitForIBDStart()
	defer flow.FinishIBD()

	peerSelectedTipHash := flow.peer.SelectedTipHash()
	highestSharedBlockHash, err := flow.findHighestSharedBlockHash(peerSelectedTipHash)
	if err != nil {
		return err
	}
	if flow.DAG().IsKnownFinalizedBlock(highestSharedBlockHash) {
		return protocolerrors.Errorf(false, "cannot initiate "+
			"IBD with peer %s because the highest shared chain block (%s) is "+
			"below the finality point", flow.peer, highestSharedBlockHash)
	}

	return flow.downloadBlocks(highestSharedBlockHash, peerSelectedTipHash)
}

func (flow *handleIBDFlow) findHighestSharedBlockHash(peerSelectedTipHash *daghash.Hash) (lowHash *daghash.Hash,
	err error) {

	lowHash = flow.DAG().Params.GenesisHash
	highHash := peerSelectedTipHash

	for {
		err := flow.sendGetBlockLocator(lowHash, highHash)
		if err != nil {
			return nil, err
		}

		blockLocatorHashes, err := flow.receiveBlockLocator()
		if err != nil {
			return nil, err
		}

		// We check whether the locator's highest hash is in the local DAG.
		// If it is, return it. If it isn't, we need to narrow our
		// getBlockLocator request and try again.
		locatorHighHash := blockLocatorHashes[0]
		if flow.DAG().IsInDAG(locatorHighHash) {
			return locatorHighHash, nil
		}

		highHash, lowHash = flow.DAG().FindNextLocatorBoundaries(blockLocatorHashes)
	}
}

func (flow *handleIBDFlow) sendGetBlockLocator(lowHash *daghash.Hash, highHash *daghash.Hash) error {

	msgGetBlockLocator := domainmessage.NewMsgRequestBlockLocator(highHash, lowHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockLocator)
}

func (flow *handleIBDFlow) receiveBlockLocator() (blockLocatorHashes []*daghash.Hash, err error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgBlockLocator, ok := message.(*domainmessage.MsgBlockLocator)
	if !ok {
		return nil,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", domainmessage.CmdBlockLocator, message.Command())
	}
	return msgBlockLocator.BlockLocatorHashes, nil
}

func (flow *handleIBDFlow) downloadBlocks(highestSharedBlockHash *daghash.Hash,
	peerSelectedTipHash *daghash.Hash) error {

	err := flow.sendGetBlocks(highestSharedBlockHash, peerSelectedTipHash)
	if err != nil {
		return err
	}

	blocksReceived := 0
	for {
		msgIBDBlock, doneIBD, err := flow.receiveIBDBlock()
		if err != nil {
			return err
		}

		if doneIBD {
			return nil
		}

		err = flow.processIBDBlock(msgIBDBlock)
		if err != nil {
			return err
		}

		blocksReceived++
		if blocksReceived%ibdBatchSize == 0 {
			err = flow.outgoingRoute.Enqueue(domainmessage.NewMsgRequestNextIBDBlocks())
			if err != nil {
				return err
			}
		}
	}
}

func (flow *handleIBDFlow) sendGetBlocks(highestSharedBlockHash *daghash.Hash,
	peerSelectedTipHash *daghash.Hash) error {

	msgGetBlockInvs := domainmessage.NewMsgRequstIBDBlocks(highestSharedBlockHash, peerSelectedTipHash)
	return flow.outgoingRoute.Enqueue(msgGetBlockInvs)
}

func (flow *handleIBDFlow) receiveIBDBlock() (msgIBDBlock *domainmessage.MsgIBDBlock, doneIBD bool, err error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}
	switch message := message.(type) {
	case *domainmessage.MsgIBDBlock:
		return message, false, nil
	case *domainmessage.MsgDoneIBDBlocks:
		return nil, true, nil
	default:
		return nil, false,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", domainmessage.CmdIBDBlock, message.Command())
	}
}

func (flow *handleIBDFlow) processIBDBlock(msgIBDBlock *domainmessage.MsgIBDBlock) error {

	block := util.NewBlock(msgIBDBlock.MsgBlock)
	if flow.DAG().IsInDAG(block.Hash()) {
		return nil
	}
	isOrphan, isDelayed, err := flow.DAG().ProcessBlock(block, blockdag.BFNone)
	if err != nil {
		return err
	}
	if isOrphan {
		return protocolerrors.Errorf(true, "received orphan block %s "+
			"during IBD", block.Hash())
	}
	if isDelayed {
		return protocolerrors.Errorf(false, "received delayed block %s "+
			"during IBD", block.Hash())
	}
	err = flow.OnNewBlock(block)
	if err != nil {
		return err
	}
	return nil
}
