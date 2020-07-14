package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
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

			shouldContinue, err := sendGetBlockLocator(outgoingRoute, peer, dag)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}
}

func sendGetBlockLocator(outgoingRoute *router.Route, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG) (shouldContinue bool, err error) {

	selectedTipHash, err := peer.SelectedTipHash()
	if err != nil {
		return false, err
	}
	msgGetBlockLocator := wire.NewMsgGetBlockLocator(selectedTipHash, dag.Params.GenesisHash)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgGetBlockLocator, common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	return isOpen, nil
}

func finishIBD() {
	atomic.StoreUint32(&isIBDRunning, 0)
}
