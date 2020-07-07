package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/messagemux"
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/protocol/blockrelay"
	"github.com/kaspanet/kaspad/protocol/getrelayblockslistener"
	"github.com/kaspanet/kaspad/wire"
	"sync/atomic"
)

// StartProtocol starts the p2p protocol for a given connection
func StartProtocol(server p2pserver.Server, mux messagemux.Mux, connection p2pserver.Connection,
	dag *blockdag.BlockDAG) {

	stop := make(chan struct{})
	shutdown := uint32(0)

	blockRelayCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdInvRelayBlock, wire.CmdBlock}, blockRelayCh)
	spawn(func() {
		err := blockrelay.StartBlockRelay(blockRelayCh, server, connection, dag)
		if err == nil {
			return
		}

		log.Errorf("error from StartBlockRelay: %s", err)
		if atomic.LoadUint32(&shutdown) == 0 {
			stop <- struct{}{}
		}
	})

	getRelayBlocksListenerCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdGetRelayBlocks}, getRelayBlocksListenerCh)
	spawn(func() {
		err := getrelayblockslistener.StartGetRelayBlocksListener(getRelayBlocksListenerCh, connection, dag)
		if err == nil {
			return
		}

		log.Errorf("error from StartGetRelayBlocksListener: %s", err)
		if atomic.LoadUint32(&shutdown) == 0 {
			stop <- struct{}{}
		}
	})

	<-stop
	atomic.StoreUint32(&shutdown, 1)
}
