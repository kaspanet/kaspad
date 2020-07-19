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

// StartIBDIfRequired selects a peer and starts IBD against it
// if required
func StartIBDIfRequired(dag *blockdag.BlockDAG) {
	startIBDMutex.Lock()
	defer startIBDMutex.Unlock()

	if atomic.LoadUint32(&isIBDRunning) != 0 {
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
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) error {

	for {
		peer.WaitForIBDStart()

		// We run the flow inside a func so that the defer is called at its end
		shouldStop, err := func() (bool, error) {
			defer finishIBD(dag)

			peerSelectedTipHash := peer.SelectedTipHash()
			highestSharedBlockHash, shouldStop, err := findHighestSharedBlockHash(incomingRoute, outgoingRoute, dag, peerSelectedTipHash)
			if err != nil {
				return true, err
			}
			if shouldStop {
				return true, nil
			}
			if dag.IsKnownFinalizedBlock(highestSharedBlockHash) {
				return true, protocolerrors.Errorf(false, "cannot initiate "+
					"IBD with peer %s because the highest shared chain block (%s) is "+
					"below the finality point", peer, highestSharedBlockHash)
			}

			shouldStop, err = downloadBlocks(incomingRoute, outgoingRoute, dag, highestSharedBlockHash, peerSelectedTipHash)
			if err != nil {
				return true, err
			}
			return shouldStop, nil
		}()
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
}

func findHighestSharedBlockHash(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG,
	peerSelectedTipHash *daghash.Hash) (lowHash *daghash.Hash, shouldStop bool, err error) {

	lowHash = dag.Params.GenesisHash
	highHash := peerSelectedTipHash

	for {
		shouldStop, err = sendGetBlockLocator(outgoingRoute, lowHash, highHash)
		if err != nil {
			return nil, true, err
		}
		if shouldStop {
			return nil, true, nil
		}

		blockLocatorHashes, shouldStop, err := receiveBlockLocator(incomingRoute)
		if err != nil {
			return nil, true, err
		}
		if shouldStop {
			return nil, true, nil
		}

		// We check whether the locator's highest hash is in the local DAG.
		// If it is, return it. If it isn't, we need to narrow our
		// getBlockLocator request and try again.
		locatorHighHash := blockLocatorHashes[0]
		if dag.IsInDAG(locatorHighHash) {
			return locatorHighHash, false, nil
		}

		highHash, lowHash = dag.FindNextLocatorBoundaries(blockLocatorHashes)
	}
}

func sendGetBlockLocator(outgoingRoute *router.Route, lowHash *daghash.Hash,
	highHash *daghash.Hash) (shouldStop bool, err error) {

	msgGetBlockLocator := wire.NewMsgGetBlockLocator(highHash, lowHash)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgGetBlockLocator, common.DefaultTimeout)
	if err != nil {
		return true, err
	}
	return !isOpen, nil
}

func receiveBlockLocator(incomingRoute *router.Route) (blockLocatorHashes []*daghash.Hash,
	shouldStop bool, err error) {

	message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, true, err
	}
	if !isOpen {
		return nil, true, nil
	}
	msgBlockLocator, ok := message.(*wire.MsgBlockLocator)
	if !ok {
		return nil, true,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: MsgBlockLocator, got: %s", message.Command())
	}
	return msgBlockLocator.BlockLocatorHashes, false, nil
}

func downloadBlocks(incomingRoute *router.Route, outgoingRoute *router.Route,
	dag *blockdag.BlockDAG, highestSharedBlockHash *daghash.Hash, peerSelectedTipHash *daghash.Hash) (shouldStop bool, err error) {

	shouldStop, err = sendGetBlocks(outgoingRoute, highestSharedBlockHash, peerSelectedTipHash)
	if err != nil {
		return true, err
	}
	if shouldStop {
		return true, nil
	}

	for {
		msgIBDBlock, shouldStop, err := receiveIBDBlock(incomingRoute)
		if err != nil {
			return true, err
		}
		if shouldStop {
			return true, nil
		}
		shouldStop, err = processIBDBlock(dag, msgIBDBlock)
		if err != nil {
			return true, err
		}
		if shouldStop {
			return true, nil
		}
		if msgIBDBlock.BlockHash().IsEqual(peerSelectedTipHash) {
			return true, nil
		}
	}
}

func sendGetBlocks(outgoingRoute *router.Route, highestSharedBlockHash *daghash.Hash,
	peerSelectedTipHash *daghash.Hash) (shouldStop bool, err error) {

	msgGetBlockInvs := wire.NewMsgGetBlocks(highestSharedBlockHash, peerSelectedTipHash)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgGetBlockInvs, common.DefaultTimeout)
	if err != nil {
		return true, err
	}
	return !isOpen, nil
}

func receiveIBDBlock(incomingRoute *router.Route) (msgIBDBlock *wire.MsgIBDBlock, shouldStop bool, err error) {
	message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, true, err
	}
	if !isOpen {
		return nil, true, nil
	}
	msgIBDBlock, ok := message.(*wire.MsgIBDBlock)
	if !ok {
		return nil, true,
			protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: MsgIBDBlock, got: %s", message.Command())
	}
	return msgIBDBlock, false, nil
}

func processIBDBlock(dag *blockdag.BlockDAG, msgIBDBlock *wire.MsgIBDBlock) (shouldStop bool, err error) {
	block := util.NewBlock(&msgIBDBlock.MsgBlock)
	if dag.IsInDAG(block.Hash()) {
		return false, nil
	}
	isOrphan, isDelayed, err := dag.ProcessBlock(block, blockdag.BFNone)
	if err != nil {
		return true, err
	}
	if isOrphan {
		return true, protocolerrors.Errorf(true, "received orphan block %s "+
			"during IBD", block.Hash())
	}
	if isDelayed {
		// Theoretically we may receive delayed blocks during IBD from
		// an honest peer, so we don't bad/disconnect them
		return true, nil
	}
	return false, nil
}

func finishIBD(dag *blockdag.BlockDAG) {
	atomic.StoreUint32(&isIBDRunning, 0)

	StartIBDIfRequired(dag)
}
