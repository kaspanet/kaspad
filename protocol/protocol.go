package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/wire"
)

// ProtocolManager manages the p2p protocol
type ProtocolManager struct {
	netAdapter *netadapter.NetAdapter
}

// Start starts the p2p protocol manager
func Start(listeningPort string, dag *blockdag.BlockDAG) (*ProtocolManager, error) {
	netAdapter, err := netadapter.NewNetAdapter(listeningPort)
	if err != nil {
		return nil, err
	}

	routerInitializer := buildRouterInitializer(netAdapter, dag)
	netAdapter.SetRouterInitializer(routerInitializer)

	protocolManager := ProtocolManager{
		netAdapter: netAdapter,
	}
	return &protocolManager, nil
}

func buildRouterInitializer(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) func() *netadapter.Router {
	return func() *netadapter.Router {
		router := netadapter.Router{}
		router.AddRoute([]string{wire.CmdTx}, startDummy(netAdapter, dag))
		return &router
	}
}

func startDummy(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) chan<- wire.Message {
	ch := make(chan wire.Message)
	spawn(func() {
		for range ch {
		}
	})
	return ch
}
