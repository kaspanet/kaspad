package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
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

		// We the flow inside a func so that the defer is called at its end
		err := func() error {
			defer finishIBD()

			return nil
		}()
		if err != nil {
			return err
		}
	}
}

func finishIBD() {
	atomic.StoreUint32(&isIBDRunning, 0)
}
