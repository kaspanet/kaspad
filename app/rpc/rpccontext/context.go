package rpccontext

import (
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Context struct {
	NetAdapter             *netadapter.NetAdapter
	DAG                    *blockdag.BlockDAG
	ProtocolManager        *protocol.Manager
	ConnectionManager      *connmanager.ConnectionManager
	BlockTemplateGenerator *mining.BlkTmplGenerator
	Mempool                *mempool.TxPool

	BlockTemplateState *BlockTemplateState
}

func NewContext(
	netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	blockTemplateGenerator *mining.BlkTmplGenerator,
	mempool *mempool.TxPool) *Context {
	context := &Context{
		NetAdapter:             netAdapter,
		DAG:                    dag,
		ProtocolManager:        protocolManager,
		ConnectionManager:      connectionManager,
		BlockTemplateGenerator: blockTemplateGenerator,
		Mempool:                mempool,
	}
	context.BlockTemplateState = NewBlockTemplateState(context)
	return context
}
