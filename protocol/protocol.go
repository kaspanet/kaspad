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
		err := router.AddRoute([]string{wire.CmdPing, wire.CmdPong}, startPing(netAdapter, dag))
		if err != nil {
			return nil, err
		}
		return router, nil
	}
}

// TODO(libp2p): Remove this and change it with a real Ping-Pong flow.
func startPing(netAdapter *netadapter.NetAdapter, router *netadapter.Router,
	dag *blockdag.BlockDAG) chan wire.Message {

	ch := make(chan wire.Message)
	spawn(func() {
		router.WriteOutgoingMessage(wire.NewMsgPing(666))
		for message := range ch {
			log.Infof("Got message: %+v", message.Command())
			if message.Command() == "ping" {
				router.WriteOutgoingMessage(wire.NewMsgPong(666))
			}
		}
	})
	return route
}
