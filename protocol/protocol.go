package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/messagemux"
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/protocol/blockrelay"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"sync/atomic"
)

// StartProtocol starts the p2p protocol for a given connection
func StartProtocol(server p2pserver.Server, mux messagemux.Mux, connection p2pserver.Connection,
	dag *blockdag.BlockDAG, requestedBlocks *blockrelay.SharedRequestedBlocks) {

	stop := make(chan struct{})
	shutdown := uint32(0)

	blockRelayCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdTx}, blockRelayCh)
	spawn(func() {
		err := blockrelay.StartBlockRelay(blockRelayCh, server, connection, dag, requestedBlocks)
		if err == nil {
			return
		}

		log.Errorf("error from StartBlockRelay: %s", err)
		if atomic.LoadUint32(&shutdown) == 0 {
			stop <- struct{}{}
		}
	})

	<-stop
	atomic.StoreUint32(&shutdown, 1)
}

func AddBanScoreAndPushRejectMsg(connection p2pserver.Connection, command string, code wire.RejectCode, hash *daghash.Hash, persistent, transient uint32, reason string) (isBanned bool) {
	PushRejectMsg(connection, command, code, reason, hash)
	return connection.AddBanScore(persistent, transient, reason)
}

func PushRejectMsg(connection p2pserver.Connection, command string, code wire.RejectCode, reason string, hash *daghash.Hash) {
	msg := wire.NewMsgReject(command, code, reason)
	err := connection.Send(msg)
	if err != nil {
		log.Errorf("couldn't send reject message to %s", connection)
	}
}
