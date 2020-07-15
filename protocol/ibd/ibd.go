package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"sync"
	"sync/atomic"
)

var (
	isIBDRunning  uint32
	startIBDMutex sync.Mutex
)

func StartIBDIfRequired(dag *blockdag.BlockDAG) error {
	startIBDMutex.Lock()
	defer startIBDMutex.Unlock()

	if atomic.LoadUint32(&isIBDRunning) != 0 {
		return nil
	}

	peer := selectPeerForIBD()
	if peer == nil {
		return requestSelectedTipsIfRequired(dag)
	}

	atomic.StoreUint32(&isIBDRunning, 1)
	peer.StartIBD()
	return nil
}

func selectPeerForIBD() *peerpkg.Peer {
	return nil
}

func requestSelectedTipsIfRequired(dag *blockdag.BlockDAG) error {
	if recentlyReceivedBlock(dag) {
		return nil
	}
	return requestSelectedTips(dag)
}

func recentlyReceivedBlock(dag *blockdag.BlockDAG) bool {
	return false
}

func requestSelectedTips(dag *blockdag.BlockDAG) error {
	return nil
}

func HandleIBD(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) error {

	for {
		peer.WaitForIBDStart()

		// We run the flow inside a func so that the defer is called at its end
		err := func() error {
			defer finishIBD()

			lowHash, shouldContinue, err := findIBDLowHash(incomingRoute, outgoingRoute, peer, dag)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
			if dag.IsKnownFinalizedBlock(lowHash) {
				return protocolerrors.Errorf(false, "Cannot initiate "+
					"IBD with peer %s because the highest shared chain block (%s) is "+
					"below the finality point", peer, lowHash)
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}
}

func findIBDLowHash(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) (lowHash *daghash.Hash, shouldContinue bool, err error) {

	lowHash = dag.Params.GenesisHash
	highHash, err := peer.SelectedTipHash()
	if err != nil {
		return nil, false, err
	}

	for {
		shouldContinue, err = sendGetBlockLocator(outgoingRoute, lowHash, highHash)
		if err != nil {
			return nil, false, err
		}
		if !shouldContinue {
			return nil, false, nil
		}

		blockLocatorHashes, shouldContinue, err := receiveBlockLocator(incomingRoute)
		if err != nil {
			return nil, false, err
		}
		if !shouldContinue {
			return nil, false, nil
		}

		// We check whether the locator's highest hash is in the local DAG.
		// If it isn't, we need to narrow our getBlockLocator request and
		// try again.
		locatorHighHash := blockLocatorHashes[0]
		if !dag.IsInDAG(locatorHighHash) {
			highHash, lowHash = dag.FindNextLocatorBoundaries(blockLocatorHashes)
			continue
		}

		// We return the locator's highest hash as the lowHash here.
		// This is not a mistake. The blocks we desire start from the highest
		// hash that we know of and end at the highest hash that the peer
		// knows of (i.e. its selected tip).
		return locatorHighHash, true, nil
	}
}

func sendGetBlockLocator(outgoingRoute *router.Route, lowHash *daghash.Hash,
	highHash *daghash.Hash) (shouldContinue bool, err error) {

	msgGetBlockLocator := wire.NewMsgGetBlockLocator(highHash, lowHash)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgGetBlockLocator, common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	return isOpen, nil
}

func receiveBlockLocator(incomingRoute *router.Route) (blockLocatorHashes []*daghash.Hash,
	shouldContinue bool, err error) {

	message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}
	if !isOpen {
		return nil, false, nil
	}
	msgBlockLocator, ok := message.(*wire.MsgBlockLocator)
	if !ok {
		return nil, false, protocolerrors.Errorf(true, "unexpected message")
	}
	return msgBlockLocator.BlockLocatorHashes, true, nil
}

func finishIBD() {
	atomic.StoreUint32(&isIBDRunning, 0)
}
