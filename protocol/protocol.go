package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/messagemux"
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/wire"
)

// StartProtocol starts the p2p protocol for a given connection
func StartProtocol(server p2pserver.Server, mux messagemux.Mux, connection p2pserver.Connection, dag *blockdag.BlockDAG) {
	mux.AddFlow([]string{wire.CmdTx}, startDummy(server, connection, dag))
}

func startDummy(server p2pserver.Server, connection p2pserver.Connection, dag *blockdag.BlockDAG) chan<- wire.Message {
	ch := make(chan wire.Message)
	go func() {
		for range ch {
		}
	}()
	return ch
}
