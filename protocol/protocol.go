package protocol

import (
	"fmt"
	"sync/atomic"

	"github.com/kaspanet/kaspad/addressmanager"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/flows/addressexchange"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/protocol/flows/handshake"
	"github.com/kaspanet/kaspad/protocol/flows/ibd"
	"github.com/kaspanet/kaspad/protocol/flows/ibd/selectedtip"
	"github.com/kaspanet/kaspad/protocol/flows/ping"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (m *Manager) routerInitializer(router *routerpkg.Router, netConnection *netadapter.NetConnection) {
	spawn("routerInitializer-startFlows", func() {
		isBanned, err := m.context.ConnectionManager().IsBanned(netConnection)
		if err != nil && !errors.Is(err, addressmanager.ErrAddressNotFound) {
			panic(err)
		}
		if isBanned {
			netConnection.Disconnect()
			return
		}

		err = m.startFlows(netConnection, router)
		if err != nil {
			if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
				if protocolErr.ShouldBan {
					err := m.context.ConnectionManager().Ban(netConnection)
					if err != nil && !errors.Is(err, addressmanager.ErrAddressNotFound) {
						panic(err)
					}
				}
				netConnection.Disconnect()
				return
			}
			if errors.Is(err, routerpkg.ErrTimeout) {
				netConnection.Disconnect()
				return
			}
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				return
			}
			panic(err)
		}
	})
}

func (m *Manager) startFlows(netConnection *netadapter.NetConnection, router *routerpkg.Router) error {
	stop := make(chan error)
	isStopping := uint32(0)

	netConnection.SetOnInvalidMessageHandler(func(err error) {
		if atomic.AddUint32(&isStopping, 1) == 1 {
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

	m.addAddressFlows(router, &isStopping, stop, peer)
	m.addBlockRelayFlows(router, &isStopping, stop, peer)
	m.addPingFlows(router, &isStopping, stop, peer)
	m.addIBDFlows(router, &isStopping, stop, peer)
	m.addTransactionRelayFlow(router, &isStopping, stop)

	err = <-stop
	return err
}

func (m *Manager) addAddressFlows(router *routerpkg.Router, isStopping *uint32, stop chan error,
	peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	m.addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return addressexchange.SendAddresses(m.context, incomingRoute, outgoingRoute)
		},
	)

	m.addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return addressexchange.ReceiveAddresses(m.context, incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addBlockRelayFlows(router *routerpkg.Router, isStopping *uint32, stop chan error, peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	m.addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayInvs(m.context, incomingRoute,
				outgoingRoute, peer)
		},
	)

	m.addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return blockrelay.HandleRelayBlockRequests(m.context, incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addPingFlows(router *routerpkg.Router, isStopping *uint32, stop chan error, peer *peerpkg.Peer) {
	outgoingRoute := router.OutgoingRoute()

	m.addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.ReceivePings(m.context, incomingRoute, outgoingRoute)
		},
	)

	m.addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ping.SendPings(m.context, incomingRoute, outgoingRoute, peer)
		},
	)
}

func (m *Manager) addIBDFlows(router *routerpkg.Router, isStopping *uint32, stop chan error,
	peer *peerpkg.Peer) {

	outgoingRoute := router.OutgoingRoute()

	m.addFlow("HandleIBD", router, []wire.MessageCommand{wire.CmdBlockLocator, wire.CmdIBDBlock}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleIBD(m.context, incomingRoute, outgoingRoute, peer)
		},
	)

	m.addFlow("RequestSelectedTip", router, []wire.MessageCommand{wire.CmdSelectedTip}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return selectedtip.RequestSelectedTip(m.context, incomingRoute, outgoingRoute, peer)
		},
	)

	m.addFlow("HandleGetSelectedTip", router, []wire.MessageCommand{wire.CmdGetSelectedTip}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return selectedtip.HandleGetSelectedTip(m.context, incomingRoute, outgoingRoute)
		},
	)

	m.addFlow("HandleGetBlockLocator", router, []wire.MessageCommand{wire.CmdGetBlockLocator}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlockLocator(m.context, incomingRoute, outgoingRoute)
		},
	)

	m.addFlow("HandleGetBlocks", router, []wire.MessageCommand{wire.CmdGetBlocks}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return ibd.HandleGetBlocks(m.context, incomingRoute, outgoingRoute)
		},
	)
}

func (m *Manager) addTransactionRelayFlow(router *routerpkg.Router, isStopping *uint32, stop chan error) {

	outgoingRoute := router.OutgoingRoute()

	m.addFlow("HandleRelayedTransactions", router, []wire.MessageCommand{wire.CmdInv, wire.CmdTx}, isStopping, stop,
		func(incomingRoute *routerpkg.Route) error {
			return relaytransactions.HandleRelayedTransactions(m.context, incomingRoute, outgoingRoute)
		},
	)
}

func (m *Manager) addFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, isStopping *uint32,
	stopChan chan error, flow func(route *routerpkg.Route) error) {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	spawn(fmt.Sprintf("addFlow-startFlow-%s", name), func() {
		err := flow(route)
		if err != nil {
			m.context.HandleError(err, name, isStopping, stopChan)
			return
		}
	})
}

func (m *Manager) addOneTimeFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand,
	isStopping *uint32, stopChan chan error, flow func(route *routerpkg.Route) error) {

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
			m.context.HandleError(err, name, isStopping, stopChan)
			return
		}
	})
}
