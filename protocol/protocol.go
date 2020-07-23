package protocol

import (
	"fmt"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol/flows/ibd/selectedtip"
	"sync/atomic"

	"github.com/kaspanet/kaspad/protocol/flows/handshake"

	"github.com/kaspanet/kaspad/protocol/flows/addressexchange"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"

	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/flows/ibd"
	"github.com/kaspanet/kaspad/protocol/flows/ping"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (m *Manager) routerInitializer(netConnection *netadapter.NetConnection) (*routerpkg.Router, error) {
	router := routerpkg.NewRouter()
	spawn("newRouterInitializer-startFlows", func() {
		err := m.startFlows(netConnection, router)
		if err != nil {
			if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
				if protocolErr.ShouldBan {
					m.context.ConnectionManager().Ban(netConnection)
				}
				err = m.context.NetAdapter().Disconnect(netConnection)
				if err != nil {
					panic(err)
				}
				return
			}
			if errors.Is(err, routerpkg.ErrTimeout) {
				err = m.context.NetAdapter().Disconnect(netConnection)
				if err != nil {
					panic(err)
				}
				return
			}
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				return
			}
			panic(err)
		}
	})
	return router, nil
}

func (m *Manager) startFlows(netConnection *netadapter.NetConnection, router *routerpkg.Router) error {
	stop := make(chan error)
	stopped := uint32(0)

	netConnection.SetOnInvalidMessageHandler(func(err error) {
		if atomic.AddUint32(&stopped, 1) == 1 {
			stop <- protocolerrors.Wrap(true, err, "received bad message")
		}
	})

	peer, closed, err := handshake.HandleHandshake(m.context, router, netConnection)
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	m.addAddressFlows(router, &stopped, stop, peer)
	m.addBlockRelayFlows(router, &stopped, stop, peer)
	m.addPingFlows(router, &stopped, stop, peer)
	m.addIBDFlows(router, &stopped, stop, peer)
	m.addTransactionRelayFlow(router, &stopped, stop)

	err = <-stop
	return err
}

func (m *Manager) addAddressFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return addressexchange.SendAddresses(m.context, incomingRoute, outgoingRoute)
		},
	)

	addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return addressexchange.ReceiveAddresses(m.context, incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addBlockRelayFlows(router *routerpkg.Router, stopped *uint32, stop chan error, peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayInvs(m.context, incomingRoute,
				outgoingRoute, peer)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayBlockRequests(m.context, incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addPingFlows(router *routerpkg.Router, stopped *uint32, stop chan error, peer *peerpkg.Peer) {
	outgoingRoute := router.OutgoingRoute()

	addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(m.context, incomingRoute, outgoingRoute)
		},
	)

	addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(m.context, incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addIBDFlows(router *routerpkg.Router, stopped *uint32, stop chan error,
	peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleIBD", router, []wire.MessageCommand{wire.CmdBlockLocator, wire.CmdIBDBlock}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleIBD(m.context, incomingRoute, outgoingRoute, peer)
		},
	)

	addFlow("RequestSelectedTip", router, []wire.MessageCommand{wire.CmdSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return selectedtip.RequestSelectedTip(m.context, incomingRoute, outgoingRoute, peer)
		},
	)

	addFlow("HandleGetSelectedTip", router, []wire.MessageCommand{wire.CmdGetSelectedTip}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return selectedtip.HandleGetSelectedTip(m.context, incomingRoute, outgoingRoute)
		},
	)

	addFlow("HandleGetBlockLocator", router, []wire.MessageCommand{wire.CmdGetBlockLocator}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlockLocator(m.context, incomingRoute, outgoingRoute)
		},
	)

	addFlow("HandleGetBlocks", router, []wire.MessageCommand{wire.CmdGetBlocks}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlocks(m.context, incomingRoute, outgoingRoute)
		},
	)
}

func (m *Manager) addTransactionRelayFlow(router *routerpkg.Router, stopped *uint32, stop chan error) {

	outgoingRoute := router.OutgoingRoute()

	addFlow("HandleRelayedTransactions", router, []wire.MessageCommand{wire.CmdInv, wire.CmdTx}, stopped, stop,
		func(incomingRoute *routerpkg.Route) error {
			return relaytransactions.HandleRelayedTransactions(m.context, incomingRoute, outgoingRoute)
		},
	)
}

func addFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, stopped *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(fmt.Sprintf("addFlow-startFlow-%s", name), func() {
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
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(fmt.Sprintf("addOneTimeFlow-startFlow-%s", name), func() {
		defer func() {
			err := router.RemoveRoute(messageTypes)
			if err != nil {
				panic(err)
			}
		}()

		err := flow(route)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if err != nil && atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
}
