package rpc

import (
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/util"
)

type Manager struct {
	context *rpccontext.Context
}

func NewManager(
	netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	blockTemplateGenerator *mining.BlkTmplGenerator,
	mempool *mempool.TxPool,
	addressManager *addressmanager.AddressManager) *Manager {
	manager := Manager{
		context: rpccontext.NewContext(
			netAdapter,
			dag,
			protocolManager,
			connectionManager,
			blockTemplateGenerator,
			mempool,
			addressManager,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}

func (m *Manager) NotifyBlockAddedToDAG(block *util.Block) {
	m.context.BlockTemplateState.NotifyBlockAdded(block)
	m.context.NotificationManager.NotifyBlockAdded(block)
}

func (m *Manager) NotifyTransactionAddedToMempool() {
	m.context.BlockTemplateState.NotifyMempoolTx()
}
