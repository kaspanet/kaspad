package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/handlerelayblockrequests"
	"github.com/kaspanet/kaspad/protocol/handlerelayinvs"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
	"sync/atomic"
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
	return func() (*routerpkg.Router, error) {
		router := routerpkg.NewRouter()
		spawn(func() {
			err := startFlows(netAdapter, router, dag)
			if err != nil {
				// TODO(libp2p) Ban peer
			}
		})
		return router, nil
	}
}

func startFlows(netAdapter *netadapter.NetAdapter, router *routerpkg.Router, dag *blockdag.BlockDAG) error {
	stop := make(chan error)
	stopped := uint32(0)

	outgoingRoute := router.OutgoingRoute()
	peer := new(peerpkg.Peer)

	addFlow("HandleRelayInvs", router, []string{wire.CmdInvRelayBlock, wire.CmdBlock}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return handlerelayinvs.HandleRelayInvs(incomingRoute, peer, netAdapter, outgoingRoute, dag)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []string{wire.CmdGetRelayBlocks}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return handlerelayblockrequests.HandleRelayBlockRequests(incomingRoute, peer, outgoingRoute, dag)
		},
	)

	// TODO(libp2p): Remove this and change it with a real Ping-Pong flow.
	addFlow("PingPong", router, []string{wire.CmdPing, wire.CmdPong}, &stopped, stop, func(incomingRoute *routerpkg.Route) error {
		err := outgoingRoute.Enqueue(wire.NewMsgPing(666))
		if err != nil {
			return err
		}
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		for {
			log.Infof("Got message: %+v", message.Command())
			if message.Command() == "ping" {
				err := outgoingRoute.Enqueue(wire.NewMsgPong(666))
				if err != nil {
					return err
				}
			}
		}
	})

	err := <-stop
	return err
}

func addFlow(name string, router *routerpkg.Router, messageTypes []string, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route := routerpkg.NewRoute()
	err := router.AddRoute(messageTypes, route)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		err := flow(route)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
}
