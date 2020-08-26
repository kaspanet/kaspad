package rpc

import "github.com/kaspanet/kaspad/infrastructure/network/netadapter"

type Manager struct {
	netAdapter *netadapter.NetAdapter
}

func NewManager(netAdapter *netadapter.NetAdapter) *Manager {
	manager := Manager{
		netAdapter: netAdapter,
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}
