package protocol

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/app/protocol/flows/rejects"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/addressexchange"
	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/app/protocol/flows/handshake"
	"github.com/kaspanet/kaspad/app/protocol/flows/ping"
	"github.com/kaspanet/kaspad/app/protocol/flows/transactionrelay"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
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
			log.Infof("Peer %s is banned. Disconnecting...", netConnection)
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
			// non-blocking read from channel
			select {
			case innerError := <-errChan:
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					m.handleError(innerError, netConnection, router.OutgoingRoute())
				} else {
					log.Errorf("Peer %s sent invalid message: %s", netConnection, innerError)
					m.handleError(err, netConnection, router.OutgoingRoute())
				}
			default:
				m.handleError(err, netConnection, router.OutgoingRoute())
			}
			return
		}
		defer m.context.RemoveFromPeers(peer)

		removeHandshakeRoutes(router)

		err = m.runFlows(flows, peer, errChan)
		if err != nil {
			m.handleError(err, netConnection, router.OutgoingRoute())
			return
		}
	})
}

func (m *Manager) handleError(err error, netConnection *netadapter.NetConnection, outgoingRoute *routerpkg.Route) {
	if protocolErr := (protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
		if !m.context.Config().DisableBanning && protocolErr.ShouldBan {
			log.Warnf("Banning %s (reason: %s)", netConnection, protocolErr.Cause)

			err := m.context.ConnectionManager().Ban(netConnection)
			if !errors.Is(err, connmanager.ErrCannotBanPermanent) {
				panic(err)
			}

			err = outgoingRoute.Enqueue(appmessage.NewMsgReject(protocolErr.Error()))
			if err != nil && !errors.Is(err, routerpkg.ErrRouteClosed) {
				panic(err)
			}
		}
		log.Infof("Disconnecting from %s (reason: %s)", netConnection, protocolErr.Cause)
		netConnection.Disconnect()
		return
	}
	if errors.Is(err, routerpkg.ErrTimeout) {
		log.Warnf("Got timeout from %s. Disconnecting...", netConnection)
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
	flows = append(flows, m.registerTransactionRelayFlow(router, isStopping, errChan)...)
	flows = append(flows, m.registerRejectsFlow(router, isStopping, errChan)...)

	return flows
}

func (m *Manager) registerAddressFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("SendAddresses", router, []appmessage.MessageCommand{appmessage.CmdRequestAddresses}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.SendAddresses(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerOneTimeFlow("ReceiveAddresses", router, []appmessage.MessageCommand{appmessage.CmdAddresses}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return addressexchange.ReceiveAddresses(m.context, incomingRoute, outgoingRoute, peer)
			},
		),
	}
}

func (m *Manager) registerBlockRelayFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerOneTimeFlow("SendVirtualSelectedParentInv", router, []appmessage.MessageCommand{},
			isStopping, errChan, func(route *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.SendVirtualSelectedParentInv(m.context, outgoingRoute, peer)
			}),

		m.registerFlow("HandleRelayInvs", router, []appmessage.MessageCommand{
			appmessage.CmdInvRelayBlock, appmessage.CmdBlock, appmessage.CmdBlockLocator, appmessage.CmdIBDBlock,
			appmessage.CmdDoneHeaders, appmessage.CmdUnexpectedPruningPoint, appmessage.CmdPruningPointUTXOSetChunk,
			appmessage.CmdBlockHeaders, appmessage.CmdPruningPointHash, appmessage.CmdIBDBlockLocatorHighestHash,
			appmessage.CmdIBDBlockLocatorHighestHashNotFound, appmessage.CmdDonePruningPointUTXOSetChunks},
			isStopping, errChan, func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayInvs(m.context, incomingRoute,
					outgoingRoute, peer)
			},
		),

		m.registerFlow("HandleRelayBlockRequests", router, []appmessage.MessageCommand{appmessage.CmdRequestRelayBlocks}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRelayBlockRequests(m.context, incomingRoute, outgoingRoute, peer)
			},
		),

		m.registerFlow("HandleRequestBlockLocator", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestBlockLocator}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRequestBlockLocator(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("HandleRequestHeaders", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestHeaders, appmessage.CmdRequestNextHeaders}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRequestHeaders(m.context, incomingRoute, outgoingRoute, peer)
			},
		),

		m.registerFlow("HandleRequestPruningPointUTXOSetAndBlock", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestPruningPointUTXOSetAndBlock,
				appmessage.CmdRequestNextPruningPointUTXOSetChunk}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleRequestPruningPointUTXOSetAndBlock(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("HandleIBDBlockRequests", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestIBDBlocks}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleIBDBlockRequests(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("HandlePruningPointHashRequests", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestPruningPointHash}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandlePruningPointHashRequests(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("HandleIBDBlockLocator", router,
			[]appmessage.MessageCommand{appmessage.CmdIBDBlockLocator}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return blockrelay.HandleIBDBlockLocator(m.context, incomingRoute, outgoingRoute, peer)
			},
		),
	}
}

func (m *Manager) registerPingFlows(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("ReceivePings", router, []appmessage.MessageCommand{appmessage.CmdPing}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.ReceivePings(m.context, incomingRoute, outgoingRoute)
			},
		),

		m.registerFlow("SendPings", router, []appmessage.MessageCommand{appmessage.CmdPong}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return ping.SendPings(m.context, incomingRoute, outgoingRoute, peer)
			},
		),
	}
}

func (m *Manager) registerTransactionRelayFlow(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlowWithCapacity("HandleRelayedTransactions", 10_000, router,
			[]appmessage.MessageCommand{appmessage.CmdInvTransaction, appmessage.CmdTx, appmessage.CmdTransactionNotFound}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return transactionrelay.HandleRelayedTransactions(m.context, incomingRoute, outgoingRoute)
			},
		),
		m.registerFlow("HandleRequestTransactions", router,
			[]appmessage.MessageCommand{appmessage.CmdRequestTransactions}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return transactionrelay.HandleRequestedTransactions(m.context, incomingRoute, outgoingRoute)
			},
		),
	}
}

func (m *Manager) registerRejectsFlow(router *routerpkg.Router, isStopping *uint32, errChan chan error) []*flow {
	outgoingRoute := router.OutgoingRoute()

	return []*flow{
		m.registerFlow("HandleRejects", router,
			[]appmessage.MessageCommand{appmessage.CmdReject}, isStopping, errChan,
			func(incomingRoute *routerpkg.Route, peer *peerpkg.Peer) error {
				return rejects.HandleRejects(m.context, incomingRoute, outgoingRoute)
			},
		),
	}
}

func (m *Manager) registerFlow(name string, router *routerpkg.Router, messageTypes []appmessage.MessageCommand, isStopping *uint32,
	errChan chan error, initializeFunc flowInitializeFunc) *flow {

	route, err := router.AddIncomingRoute(messageTypes)
	if err != nil {
		panic(err)
	}

	return m.registerFlowForRoute(route, name, isStopping, errChan, initializeFunc)
}

func (m *Manager) registerFlowWithCapacity(name string, capacity int, router *routerpkg.Router,
	messageTypes []appmessage.MessageCommand, isStopping *uint32,
	errChan chan error, initializeFunc flowInitializeFunc) *flow {

	route, err := router.AddIncomingRouteWithCapacity(capacity, messageTypes)
	if err != nil {
		panic(err)
	}

	return m.registerFlowForRoute(route, name, isStopping, errChan, initializeFunc)
}

func (m *Manager) registerFlowForRoute(route *routerpkg.Route, name string, isStopping *uint32,
	errChan chan error, initializeFunc flowInitializeFunc) *flow {

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

func (m *Manager) registerOneTimeFlow(name string, router *routerpkg.Router, messageTypes []appmessage.MessageCommand,
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
	receiveVersionRoute, err := router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdVersion})
	if err != nil {
		panic(err)
	}

	sendVersionRoute, err = router.AddIncomingRoute([]appmessage.MessageCommand{appmessage.CmdVerAck})
	if err != nil {
		panic(err)
	}

	return receiveVersionRoute, sendVersionRoute
}

func removeHandshakeRoutes(router *routerpkg.Router) {
	err := router.RemoveRoute([]appmessage.MessageCommand{appmessage.CmdVersion, appmessage.CmdVerAck})
	if err != nil {
		panic(err)
	}
}
