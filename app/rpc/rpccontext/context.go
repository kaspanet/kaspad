package rpccontext

import (
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Context struct {
	NetAdapter      *netadapter.NetAdapter
	DAG             *blockdag.BlockDAG
	ProtocolManager *protocol.Manager
}

func NewContext(
	netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG,
	protocolManager *protocol.Manager) *Context {
	return &Context{
		NetAdapter:      netAdapter,
		DAG:             dag,
		ProtocolManager: protocolManager,
	}
}
