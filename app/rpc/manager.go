package rpc

import (
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Manager struct {
	context *rpccontext.Context
}

func NewManager(
	netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG,
	protocolManager *protocol.Manager) *Manager {
	manager := Manager{
		context: rpccontext.NewContext(
			netAdapter,
			dag,
			protocolManager,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}
