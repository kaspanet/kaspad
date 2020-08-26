package rpc

import (
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Manager struct {
	context *rpccontext.Context
}

func NewManager(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) *Manager {
	manager := Manager{
		context: rpccontext.NewContext(
			netAdapter,
			dag,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}
