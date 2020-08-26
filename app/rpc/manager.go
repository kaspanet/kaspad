package rpc

import (
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

type Manager struct {
	context *context
}

func NewManager(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) *Manager {
	context := newContext(netAdapter, dag)
	manager := Manager{
		context: context,
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}
