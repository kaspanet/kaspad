package protocol

import (
	"errors"
	"sync/atomic"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/handlerelayblockrequests"
	"github.com/kaspanet/kaspad/protocol/handlerelayinvs"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/ping"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/protocol/receiveaddresses"
	"github.com/kaspanet/kaspad/protocol/sendaddresses"
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

func startFlows(netAdapter *netadapter.NetAdapter, router *routerpkg.Router, dag *blockdag.BlockDAG,
	addressManager *addrmgr.AddrManager) error {
	stop := make(chan error)
	stopped := uint32(0)

	outgoingRoute := router.OutgoingRoute()
	peer := new(peerpkg.Peer)

	closed, err := handshake(router, netAdapter, peer, dag, addressManager)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return sendaddresses.SendAddresses(incomingRoute, outgoingRoute, addressManager)
		},
	)

	addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) (routeClosed bool, err error) {
			return receiveaddresses.ReceiveAddresses(incomingRoute, outgoingRoute, peer, addressManager)
		},
	)

	addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return handlerelayinvs.HandleRelayInvs(incomingRoute, outgoingRoute, peer, netAdapter, dag)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return handlerelayblockrequests.HandleRelayBlockRequests(incomingRoute, outgoingRoute, peer, dag)
		},
	)

	addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(incomingRoute, outgoingRoute)
		},
	)

	addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, &stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(incomingRoute, outgoingRoute, peer)
		},
	)

	err = <-stop
	return err
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
