package rpccontext

import (
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Context struct {
	NetAdapter        *netadapter.NetAdapter
	DAG               *blockdag.BlockDAG
	ProtocolManager   *protocol.Manager
	ConnectionManager *connmanager.ConnectionManager

	BlockTemplateGenerator *BlockTemplateGenerator
}

func NewContext(
	netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager) *Context {
	context := &Context{
		NetAdapter:        netAdapter,
		DAG:               dag,
		ProtocolManager:   protocolManager,
		ConnectionManager: connectionManager,
	}
	context.BlockTemplateGenerator = NewBlockTemplateGenerator(context)
	return context
}
