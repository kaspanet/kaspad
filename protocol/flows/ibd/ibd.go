package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

type IBDContext interface {
	DAG() *blockdag.BlockDAG
	OnNewBlock(block *util.Block) error
	StartIBDIfRequired()
	FinishIBD()
}

// HandleIBD waits for IBD start and handles it when IBD is triggered for this peer
func HandleIBD(context IBDContext, incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		err := runIBD(context, incomingRoute, outgoingRoute, peer)
		if err != nil {
			return err
		}
	}
}

func runIBD(context IBDContext, incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	peer.WaitForIBDStart()
	defer context.FinishIBD()

	peerSelectedTipHash := peer.SelectedTipHash()
	highestSharedBlockHash, err := findHighestSharedBlockHash(context, incomingRoute, outgoingRoute, peerSelectedTipHash)
	if err != nil {
		return err
	}
	if context.DAG().IsKnownFinalizedBlock(highestSharedBlockHash) {
		return protocolerrors.Errorf(false, "cannot initiate "+
			"IBD with peer %s because the highest shared chain block (%s) is "+
			"below the finality point", peer, highestSharedBlockHash)
	}

	return downloadBlocks(context, incomingRoute, outgoingRoute, highestSharedBlockHash, peerSelectedTipHash)
}

func findHighestSharedBlockHash(context IBDContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peerSelectedTipHash *daghash.Hash) (lowHash *daghash.Hash, err error) {

	lowHash = context.DAG().Params.GenesisHash
	highHash := peerSelectedTipHash

	for {
		err := sendGetBlockLocator(outgoingRoute, lowHash, highHash)
		if err != nil {
			return nil, err
		}

		blockLocatorHashes, err := receiveBlockLocator(incomingRoute)
		if err != nil {
			return nil, err
		}

		// We check whether the locator's highest hash is in the local DAG.
		// If it is, return it. If it isn't, we need to narrow our
		// getBlockLocator request and try again.
		locatorHighHash := blockLocatorHashes[0]
		if context.DAG().IsInDAG(locatorHighHash) {
			return locatorHighHash, nil
		}

		highHash, lowHash = context.DAG().FindNextLocatorBoundaries(blockLocatorHashes)
	}
}

func sendGetBlockLocator(outgoingRoute *router.Route, lowHash *daghash.Hash,
	highHash *daghash.Hash) error {

	msgGetBlockLocator := wire.NewMsgGetBlockLocator(highHash, lowHash)
	return outgoingRoute.Enqueue(msgGetBlockLocator)
}

func receiveBlockLocator(incomingRoute *router.Route) (blockLocatorHashes []*daghash.Hash, err error) {
	message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgBlockLocator, ok := message.(*wire.MsgBlockLocator)
	if !ok {
		return nil,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", wire.CmdBlockLocator, message.Command())
	}
	return msgBlockLocator.BlockLocatorHashes, nil
}

func downloadBlocks(context IBDContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	highestSharedBlockHash *daghash.Hash,
	peerSelectedTipHash *daghash.Hash) error {

	err := sendGetBlocks(outgoingRoute, highestSharedBlockHash, peerSelectedTipHash)
	if err != nil {
		return err
	}

	for {
		msgIBDBlock, err := receiveIBDBlock(incomingRoute)
		if err != nil {
			return err
		}
		err = processIBDBlock(context, msgIBDBlock)
		if err != nil {
			return err
		}
		if msgIBDBlock.BlockHash().IsEqual(peerSelectedTipHash) {
			return nil
		}
	}
}

func sendGetBlocks(outgoingRoute *router.Route, highestSharedBlockHash *daghash.Hash,
	peerSelectedTipHash *daghash.Hash) error {

	msgGetBlockInvs := wire.NewMsgGetBlocks(highestSharedBlockHash, peerSelectedTipHash)
	return outgoingRoute.Enqueue(msgGetBlockInvs)
}

func receiveIBDBlock(incomingRoute *router.Route) (msgIBDBlock *wire.MsgIBDBlock, err error) {
	message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgIBDBlock, ok := message.(*wire.MsgIBDBlock)
	if !ok {
		return nil,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", wire.CmdIBDBlock, message.Command())
	}
	return msgIBDBlock, nil
}

func processIBDBlock(context IBDContext, msgIBDBlock *wire.MsgIBDBlock) error {

	block := util.NewBlock(&msgIBDBlock.MsgBlock)
	if context.DAG().IsInDAG(block.Hash()) {
		return nil
	}
	isOrphan, isDelayed, err := context.DAG().ProcessBlock(block, blockdag.BFNone)
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
	err = context.OnNewBlock(block)
	if err != nil {
		panic(err)
	}
	return nil
}
