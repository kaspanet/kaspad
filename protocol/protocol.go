package protocol

import (
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

type flowInitializeFunc func(route *routerpkg.Route, peer *peerpkg.Peer) error
type flowExecuteFunc func(peer *peerpkg.Peer)

type flow struct {
	name        string
	executeFunc flowExecuteFunc
}

func (m *Manager) routerInitializer(router *routerpkg.Router, netConnection *netadapter.NetConnection) {
	spawn("routerInitializer-registerFlows", func() {
		isBanned, err := m.context.ConnectionManager().IsBanned(netConnection)
		if err != nil && !errors.Is(err, addressmanager.ErrAddressNotFound) {
			panic(err)
		}
		if isBanned {
			netConnection.Disconnect()
			return
		}

		isStopping := uint32(0)
		errChan := make(chan error)

		netConnection.SetOnInvalidMessageHandler(func(err error) {
			if atomic.AddUint32(&isStopping, 1) == 1 {
				errChan <- protocolerrors.Wrap(true, err, "received bad message")
			}
		})

		flows, err := m.registerFlows(router, errChan, &isStopping)
		if err != nil {
			netConnection.Disconnect()
		}

		peer, err := handshake.HandleHandshake(m.context, router, netConnection)
		if err != nil {
			m.handleError(err, netConnection)
		}

		err = m.startFlows(flows, peer, errChan)
		if err != nil {
			m.handleError(err, netConnection)
		}
	})
}

func (m *Manager) handleError(err error, netConnection *netadapter.NetConnection) {
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

func (m *Manager) registerFlows(router *routerpkg.Router, stop chan error, isStopping *uint32) (
	flows []*flow, err error) {

	flows = m.addAddressFlows(router, isStopping, stop)
	flows = append(flows, m.addBlockRelayFlows(router, isStopping, stop)...)
	flows = append(flows, m.addPingFlows(router, isStopping, stop)...)
	flows = append(flows, m.addIBDFlows(router, isStopping, stop)...)
	flows = append(flows, m.addTransactionRelayFlow(router, isStopping, stop)...)

	return flows, err
}

func (m *Manager) addAddressFlows(router *routerpkg.Router, isStopping *uint32, stop chan error) []*flow {

	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.addOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.SendAddresses(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.addOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.ReceiveAddresses(m.context, incomingRoute, outgoingRoute, peer)
			},
		)}
}

func (m *Manager) addBlockRelayFlows(router *routerpkg.Router, isStopping *uint32, stop chan error) []*flow {

	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.addFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayInvs(m.context, incomingRoute,
					outgoingRoute, peer)
			},
		),

		m.addFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayBlockRequests(m.context, incomingRoute, outgoingRoute, peer)
			},
		)}
}

func (m *Manager) addPingFlows(router *routerpkg.Router, isStopping *uint32, stop chan error) []*flow {

	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.addFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.ReceivePings(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.addFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.SendPings(m.context, incomingRoute, outgoingRoute, peer)
			},
		)}
}

func (m *Manager) addIBDFlows(router *routerpkg.Router, isStopping *uint32, stop chan error) []*flow {

	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.addFlow("HandleIBD", router, []wire.MessageCommand{wire.CmdBlockLocator, wire.CmdIBDBlock}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ibd.HandleIBD(m.context, incomingRoute, outgoingRoute, peer)
			},
		),

		m.addFlow("RequestSelectedTip", router, []wire.MessageCommand{wire.CmdSelectedTip}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return selectedtip.RequestSelectedTip(m.context, incomingRoute, outgoingRoute, peer)
			},
		),

		m.addFlow("HandleGetSelectedTip", router, []wire.MessageCommand{wire.CmdGetSelectedTip}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return selectedtip.HandleGetSelectedTip(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.addFlow("HandleGetBlockLocator", router, []wire.MessageCommand{wire.CmdGetBlockLocator}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ibd.HandleGetBlockLocator(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.addFlow("HandleGetBlocks", router, []wire.MessageCommand{wire.CmdGetBlocks}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ibd.HandleGetBlocks(m.context, incomingRoute, outgoingRoute)
			},
		)}
}

func (m *Manager) addTransactionRelayFlow(router *routerpkg.Router, isStopping *uint32, stop chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.addFlow("HandleRelayedTransactions", router, []wire.MessageCommand{wire.CmdInv, wire.CmdTx}, isStopping, stop,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return relaytransactions.HandleRelayedTransactions(m.context, incomingRoute, outgoingRoute)
			},
		)}
}

func (m *Manager) addFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, isStopping *uint32,
	stopChan chan error, initializeFunc flowInitializeFunc) *flow {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	return &flow{
		name: name,
		executeFunc: func(peer *peerpkg.Peer) {
			err := initializeFunc(route, peer)
			if err != nil {
				m.context.HandleError(err, name, isStopping, stopChan)
				return
			}
		},
	}
}

func (m *Manager) addOneTimeFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand,
	isStopping *uint32, stopChan chan error, initializeFunc flowInitializeFunc) *flow {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	return &flow{
		name: name,
		executeFunc: func(peer *peerpkg.Peer) {
			defer func() {
				err := router.RemoveRoute(messageTypes)
				if err != nil {
					panic(err)
				}
			}()

			err := initializeFunc(route, peer)
			if err != nil {
				m.context.HandleError(err, name, isStopping, stopChan)
				return
			}
		},
	}
}
