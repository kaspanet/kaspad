package protocol

import (
	"errors"
	"sync/atomic"

	"github.com/kaspanet/kaspad/protocol/flows/handshake"

	"github.com/kaspanet/kaspad/protocol/flows/addressexchange"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/flows/ping"
	"github.com/kaspanet/kaspad/protocol/ibd"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
)

// Init initializes the p2p protocol
func Init(netAdapter *netadapter.NetAdapter, addressManager *addrmgr.AddrManager, dag *blockdag.BlockDAG) {
	routerInitializer := newRouterInitializer(netAdapter, addressManager, dag)
	netAdapter.SetRouterInitializer(routerInitializer)
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

	peer, closed, err := handshake.HandleHandshake(router, netAdapter, dag, addressManager)
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

	addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return addressexchange.SendAddresses(incomingRoute, outgoingRoute, addressManager)
		},
	)

	addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return addressexchange.ReceiveAddresses(incomingRoute, outgoingRoute, peer, addressManager)
		},
	)
}

func addBlockRelayFlows(netAdapter *netadapter.NetAdapter, router *routerpkg.Router,
	stopped *uint32, stop chan error, peer *peerpkg.Peer, dag *blockdag.BlockDAG) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayInvs(incomingRoute, outgoingRoute, peer, netAdapter, dag)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayBlockRequests(incomingRoute, outgoingRoute, peer, dag)
		},
	)
}

func addPingFlows(router *routerpkg.Router, stopped *uint32, stop chan error, peer *peerpkg.Peer) {
	outgoingRoute := router.OutgoingRoute()

	addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(incomingRoute, outgoingRoute)
		},
	)

	addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(incomingRoute, outgoingRoute, peer)
		},
	)
}

func addIBDFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleIBD", router, []wire.MessageCommand{wire.CmdBlockLocator, wire.CmdIBDBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleIBD(incomingRoute, outgoingRoute, peer, dag)
		},
	)

	addFlow("RequestSelectedTip", router, []wire.MessageCommand{wire.CmdSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.RequestSelectedTip(incomingRoute, outgoingRoute, peer, dag)
		},
	)

	addFlow("HandleGetSelectedTip", router, []wire.MessageCommand{wire.CmdGetSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetSelectedTip(incomingRoute, outgoingRoute, dag)
		},
	)

	addFlow("HandleGetBlockLocator", router, []wire.MessageCommand{wire.CmdGetBlockLocator}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlockLocator(incomingRoute, outgoingRoute, dag)
		},
	)

	addFlow("HandleGetBlocks", router, []wire.MessageCommand{wire.CmdGetBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlocks(incomingRoute, outgoingRoute, dag)
		},
	)
}

func addFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, stopped *uint32,
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

func addOneTimeFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, stopped *uint32,
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
