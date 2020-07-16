package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/handlerelayblockrequests"
	"github.com/kaspanet/kaspad/protocol/handlerelayinvs"
	"github.com/kaspanet/kaspad/protocol/ibd"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/ping"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/protocol/receiveaddresses"
	"github.com/kaspanet/kaspad/protocol/sendaddresses"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"sync/atomic"
)

// Manager manages the p2p protocol
type Manager struct {
	netAdapter *netadapter.NetAdapter
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(listeningAddresses []string, dag *blockdag.BlockDAG,
	addressManager *addrmgr.AddrManager) (*Manager, error) {

	netAdapter, err := netadapter.NewNetAdapter(listeningAddresses)
	if err != nil {
		return nil, err
	}

	routerInitializer := newRouterInitializer(netAdapter, addressManager, dag)
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

func newRouterInitializer(netAdapter *netadapter.NetAdapter,
	addressManager *addrmgr.AddrManager, dag *blockdag.BlockDAG) netadapter.RouterInitializer {
	return func() (*routerpkg.Router, error) {

		router := routerpkg.NewRouter()
		spawn(func() {
			err := startFlows(netAdapter, router, dag, addressManager)
			if err != nil {
				if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
					if protocolErr.ShouldBan {
						// TODO(libp2p) Ban peer
						panic("unimplemented")
					}
					err = netAdapter.DisconnectAssociatedConnection(router)
					if err != nil {
						panic(err)
					}
					return
				}
				if errors.Is(err, routerpkg.ErrTimeout) {
					err = netAdapter.DisconnectAssociatedConnection(router)
					if err != nil {
						panic(err)
					}
					return
				}
				panic(err)
			}
		})
		return router, nil
	}
}

func startFlows(netAdapter *netadapter.NetAdapter, router *routerpkg.Router,
	dag *blockdag.BlockDAG, addressManager *addrmgr.AddrManager) error {

	stop := make(chan error)
	stopped := uint32(0)
	peer := peerpkg.New()

	closed, err := handshake(router, netAdapter, peer, dag, addressManager)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	addAddressFlows(router, &stopped, stop, peer, addressManager)
	addBlockRelayFlows(netAdapter, router, &stopped, stop, peer, dag)
	addPingFlows(router, &stopped, stop, peer)
	addIBDFlows(router, &stopped, stop, peer, dag)

	err = <-stop
	return err
}

func addAddressFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer, addressManager *addrmgr.AddrManager) {

	outgoingRoute := router.OutgoingRoute()

	addOneTimeFlow("SendAddresses", router, []string{wire.CmdGetAddresses}, stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return sendaddresses.SendAddresses(incomingRoute, outgoingRoute, addressManager)
		},
	)

	addOneTimeFlow("ReceiveAddresses", router, []string{wire.CmdAddress}, stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return receiveaddresses.ReceiveAddresses(incomingRoute, outgoingRoute, peer, addressManager)
		},
	)
}

func addBlockRelayFlows(netAdapter *netadapter.NetAdapter, router *routerpkg.Router,
	stopped *uint32, stop chan error, peer *peerpkg.Peer, dag *blockdag.BlockDAG) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleRelayInvs", router, []string{wire.CmdInvRelayBlock, wire.CmdBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return handlerelayinvs.HandleRelayInvs(incomingRoute, outgoingRoute, peer, netAdapter, dag)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []string{wire.CmdGetRelayBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return handlerelayblockrequests.HandleRelayBlockRequests(incomingRoute, outgoingRoute, peer, dag)
		},
	)
}

func addPingFlows(router *routerpkg.Router, stopped *uint32, stop chan error, peer *peerpkg.Peer) {
	outgoingRoute := router.OutgoingRoute()

	addFlow("ReceivePings", router, []string{wire.CmdPing}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(incomingRoute, outgoingRoute)
		},
	)

	addFlow("SendPings", router, []string{wire.CmdPong}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(incomingRoute, outgoingRoute, peer)
		},
	)
}

func addIBDFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleIBD", router, []string{wire.CmdBlockLocator, wire.CmdIBDBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleIBD(incomingRoute, outgoingRoute, peer, dag)
		},
	)

	addFlow("RequestSelectedTip", router, []string{wire.CmdSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.RequestSelectedTip(incomingRoute, outgoingRoute, peer, dag)
		},
	)

	addFlow("HandleGetSelectedTip", router, []string{wire.CmdGetSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetSelectedTip(incomingRoute, outgoingRoute, dag)
		},
	)

	addFlow("HandleGetBlockLocator", router, []string{wire.CmdGetBlockLocator}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlockLocator(incomingRoute, outgoingRoute, dag)
		},
	)

	addFlow("HandleGetBlocks", router, []string{wire.CmdGetBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlocks(incomingRoute, outgoingRoute, dag)
		},
	)
}

func addFlow(name string, router *routerpkg.Router, messageTypes []string, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
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

func addOneTimeFlow(name string, router *routerpkg.Router, messageTypes []string, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) (routeClosed bool, err error)) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		defer func() {
			err := router.RemoveRoute(messageTypes)
			if err != nil {
				panic(err)
			}
		}()

		closed, err := flow(route)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if (err != nil || closed) && atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
}
