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
	// isStopping flag is raised the moment that the connection associated with this router is disconnected
	// errChan is used by the flow goroutines to return to runFlows when an error occurs.
	// They are both initialized here and passed to register flows.
	isStopping := uint32(0)
	errChan := make(chan error)

	flows := m.registerFlows(router, errChan, &isStopping)
	receiveVersionRoute, sendVersionRoute := registerHandshakeRoutes(router)

	// After flows were registered - spawn a new thread that will wait for connection to finish initializing
	// and start receiving messages
	spawn("routerInitializer-runFlows", func() {
		isBanned, err := m.context.ConnectionManager().IsBanned(netConnection)
		if err != nil && !errors.Is(err, addressmanager.ErrAddressNotFound) {
			panic(err)
		}
		if isBanned {
			netConnection.Disconnect()
			return
		}

		netConnection.SetOnInvalidMessageHandler(func(err error) {
			if atomic.AddUint32(&isStopping, 1) == 1 {
				errChan <- protocolerrors.Wrap(true, err, "received bad message")
			}
		})

		peer, err := handshake.HandleHandshake(m.context, netConnection, receiveVersionRoute,
			sendVersionRoute, router.OutgoingRoute())
		if err != nil {
			m.handleError(err, netConnection)
			return
		}

		removeHandshakeRoutes(router)

		err = m.runFlows(flows, peer, errChan)
		if err != nil {
			m.handleError(err, netConnection)
			return
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

func (m *Manager) registerFlows(router *routerpkg.Router, errChan chan error, isStopping *uint32) (flows []*flow) {
	flows = m.registerAddressFlows(router, isStopping, errChan)
	flows = append(flows, m.registerBlockRelayFlows(router, isStopping, errChan)...)
	flows = append(flows, m.registerPingFlows(router, isStopping, errChan)...)
	flows = append(flows, m.registerIBDFlows(router, isStopping, errChan)...)
	flows = append(flows, m.registerTransactionRelayFlow(router, isStopping, errChan)...)

	return flows
}

func (m *Manager) registerAddressFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerOneTimeFlow("SendAddresses", router, []wire.MessageCommand{wire.CmdGetAddresses}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.SendAddresses(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerOneTimeFlow("ReceiveAddresses", router, []wire.MessageCommand{wire.CmdAddress}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.ReceiveAddresses(m.context, incomingRoute, outgoingRoute, peer)
			},
		)}
}

func (m *Manager) registerBlockRelayFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("HandleRelayInvs", router, []wire.MessageCommand{wire.CmdInvRelayBlock, wire.CmdBlock}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayInvs(m.context, incomingRoute,
					outgoingRoute, peer)
			},
		),

		m.registerFlow("HandleRelayBlockRequests", router, []wire.MessageCommand{wire.CmdGetRelayBlocks}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayBlockRequests(m.context, incomingRoute, outgoingRoute, peer)
			},
		)}
}

func (m *Manager) registerPingFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("ReceivePings", router, []wire.MessageCommand{wire.CmdPing}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.ReceivePings(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("SendPings", router, []wire.MessageCommand{wire.CmdPong}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.SendPings(m.context, incomingRoute, outgoingRoute, peer)
			},
		)}
}

func (m *Manager) registerIBDFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("HandleIBD", router, []wire.MessageCommand{wire.CmdBlockLocator, wire.CmdIBDBlock}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ibd.HandleIBD(m.context, incomingRoute, outgoingRoute, peer)
			},
		),

		m.registerFlow("RequestSelectedTip", router, []wire.MessageCommand{wire.CmdSelectedTip}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return selectedtip.RequestSelectedTip(m.context, incomingRoute, outgoingRoute, peer)
			},
		),

		m.registerFlow("HandleGetSelectedTip", router, []wire.MessageCommand{wire.CmdGetSelectedTip}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return selectedtip.HandleGetSelectedTip(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("HandleGetBlockLocator", router, []wire.MessageCommand{wire.CmdGetBlockLocator}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ibd.HandleGetBlockLocator(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("HandleGetBlocks", router, []wire.MessageCommand{wire.CmdGetBlocks}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ibd.HandleGetBlocks(m.context, incomingRoute, outgoingRoute)
			},
		)}
}

func (m *Manager) registerTransactionRelayFlow(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("HandleRelayedTransactions", router, []wire.MessageCommand{wire.CmdInv, wire.CmdTx}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return relaytransactions.HandleRelayedTransactions(m.context, incomingRoute, outgoingRoute)
			},
		)}
}

func (m *Manager) registerFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand, isStopping *uint32,
	errChan chan error, initializeFunc flowInitializeFunc) *flow {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	return &flow{
		name: name,
		executeFunc: func(peer *peerpkg.Peer) {
			err := initializeFunc(route, peer)
			if err != nil {
				m.context.HandleError(err, name, isStopping, errChan)
				return
			}
		},
	}
}

func (m *Manager) registerOneTimeFlow(name string, router *routerpkg.Router, messageTypes []wire.MessageCommand,
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

func registerHandshakeRoutes(router *routerpkg.Router) (
	receiveVersionRoute *routerpkg.Route, sendVersionRoute *routerpkg.Route) {
	receiveVersionRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVersion})
	if err != nil {
		panic(err)
	}

	sendVersionRoute, err = router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVerAck})
	if err != nil {
		panic(err)
	}

	return receiveVersionRoute, sendVersionRoute
}

func removeHandshakeRoutes(router *routerpkg.Router) {
	err := router.RemoveRoute([]wire.MessageCommand{wire.CmdVersion, wire.CmdVerAck})
	if err != nil {
		panic(err)
	}
}
