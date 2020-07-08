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

// New creates a new instance of the p2p protocol
func New(listeningAddrs []string, dag *blockdag.BlockDAG) (*ProtocolManager, error) {
	netAdapter, err := netadapter.NewNetAdapter(listeningAddrs)
	if err != nil {
		return nil, err
	}

	routerInitializer := newRouterInitializer(netAdapter, dag)
	netAdapter.SetRouterInitializer(routerInitializer)

	manager := ProtocolManager{
		netAdapter: netAdapter,
	}
	return &manager, nil
}

// Start starts the p2p protocol
func (p *ProtocolManager) Start() error {
	return p.netAdapter.Start()
}

// Stop stops the p2p protocol
func (p *ProtocolManager) Stop() error {
	return p.netAdapter.Stop()
}

func newRouterInitializer(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) netadapter.RouterInitializer {
	return func(peer *netadapter.Peer) (*netadapter.Router, error) {
		router := netadapter.Router{}
		err := router.AddRoute([]string{wire.CmdTx}, startDummy(netAdapter, peer, dag))
		if err != nil {
			return nil, err
		}
		return &router, nil
	}
}

func startDummy(netAdapter *netadapter.NetAdapter, peer *netadapter.Peer, dag *blockdag.BlockDAG) chan<- wire.Message {
	ch := make(chan wire.Message)
	spawn(func() {
		for range ch {
		}
	})
	return ch
}
