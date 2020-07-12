package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
)

// Manager manages the p2p protocol
type Manager struct {
	netAdapter *netadapter.NetAdapter
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(listeningAddrs []string, dag *blockdag.BlockDAG) (*Manager, error) {
	netAdapter, err := netadapter.NewNetAdapter(listeningAddrs)
	if err != nil {
		return nil, err
	}

	routerInitializer := newRouterInitializer(netAdapter, dag)
	netAdapter.SetRouterInitializer(routerInitializer)

	manager := Manager{
		netAdapter: netAdapter,
	}
	return &manager, nil
}

// Start starts the p2p protocol
func (p *Manager) Start() error {
	return p.netAdapter.Start()
}

// Stop stops the p2p protocol
func (p *Manager) Stop() error {
	return p.netAdapter.Stop()
}

func newRouterInitializer(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) netadapter.RouterInitializer {
	return func() (*router.Router, error) {
		router := router.NewRouter()
		err := router.AddRoute([]string{wire.CmdTx}, startDummy(netAdapter, dag))
		if err != nil {
			return nil, err
		}
		return router, nil
	}
}

func startDummy(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) *router.Route {
	route := router.NewRoute()
	spawn(func() {
		for {
			message := route.Dequeue()
			log.Infof("handling message %s", message.Command())
		}
	})
	return route
}
