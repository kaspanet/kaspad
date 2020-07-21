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
	"sync"
	"sync/atomic"
)

var (
	isIBDRunning  uint32
	startIBDMutex sync.Mutex
)

// NewBlockHandler is a function that is to be
// called when a new block is successfully processed.
type NewBlockHandler func(block *util.Block) error

// StartIBDIfRequired selects a peer and starts IBD against it
// if required
func StartIBDIfRequired(dag *blockdag.BlockDAG) {
	startIBDMutex.Lock()
	defer startIBDMutex.Unlock()

	if IsInIBD() {
		return
	}

	peer := selectPeerForIBD(dag)
	if peer == nil {
		requestSelectedTipsIfRequired(dag)
		return
	}

	atomic.StoreUint32(&isIBDRunning, 1)
	peer.StartIBD()
}

// IsInIBD is true if IBD is currently running
func IsInIBD() bool {
	return atomic.LoadUint32(&isIBDRunning) != 0
}

// selectPeerForIBD returns the first peer whose selected tip
// hash is not in our DAG
func selectPeerForIBD(dag *blockdag.BlockDAG) *peerpkg.Peer {
	for _, peer := range peerpkg.ReadyPeers() {
		peerSelectedTipHash := peer.SelectedTipHash()
		if !dag.IsInDAG(peerSelectedTipHash) {
			return peer
		}
	}
	return nil
}

// HandleIBD waits for IBD start and handles it when IBD is triggered for this peer
func HandleIBD(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG, newBlockHandler NewBlockHandler) error {

	for {
		err := runIBD(incomingRoute, outgoingRoute, peer, dag, newBlockHandler)
		if err != nil {
			return err
		}
	}
}

func runIBD(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG, newBlockHandler NewBlockHandler) error {

	peer.WaitForIBDStart()
	defer finishIBD(dag)

	peerSelectedTipHash := peer.SelectedTipHash()
	highestSharedBlockHash, err := findHighestSharedBlockHash(incomingRoute, outgoingRoute, dag, peerSelectedTipHash)
	if err != nil {
		return err
	}
	if dag.IsKnownFinalizedBlock(highestSharedBlockHash) {
		return protocolerrors.Errorf(false, "cannot initiate "+
			"IBD with peer %s because the highest shared chain block (%s) is "+
			"below the finality point", peer, highestSharedBlockHash)
	}

	return downloadBlocks(incomingRoute, outgoingRoute, dag, highestSharedBlockHash, peerSelectedTipHash,
		newBlockHandler)
}

func findHighestSharedBlockHash(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG,
	peerSelectedTipHash *daghash.Hash) (lowHash *daghash.Hash, err error) {

	lowHash = dag.Params.GenesisHash
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
		if dag.IsInDAG(locatorHighHash) {
			return locatorHighHash, nil
		}

		highHash, lowHash = dag.FindNextLocatorBoundaries(blockLocatorHashes)
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

func downloadBlocks(incomingRoute *router.Route, outgoingRoute *router.Route,
	dag *blockdag.BlockDAG, highestSharedBlockHash *daghash.Hash,
	peerSelectedTipHash *daghash.Hash, newBlockHandler NewBlockHandler) error {

	err := sendGetBlocks(outgoingRoute, highestSharedBlockHash, peerSelectedTipHash)
	if err != nil {
		return err
	}

	for {
		msgIBDBlock, err := receiveIBDBlock(incomingRoute)
		if err != nil {
			return err
		}
		err = processIBDBlock(dag, msgIBDBlock, newBlockHandler)
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

func processIBDBlock(dag *blockdag.BlockDAG, msgIBDBlock *wire.MsgIBDBlock,
	newBlockHandler NewBlockHandler) error {

	block := util.NewBlock(&msgIBDBlock.MsgBlock)
	if dag.IsInDAG(block.Hash()) {
		return nil
	}
	isOrphan, isDelayed, err := dag.ProcessBlock(block, blockdag.BFNone)
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
	err = newBlockHandler(block)
	if err != nil {
		panic(err)
	}
	return nil
}

func finishIBD(dag *blockdag.BlockDAG) {
	atomic.StoreUint32(&isIBDRunning, 0)

	StartIBDIfRequired(dag)
}
