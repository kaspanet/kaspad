package protocol

import (
	"github.com/kaspanet/kaspad/app/protocol/common"
	v3 "github.com/kaspanet/kaspad/app/protocol/flows/v3"
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flows/handshake"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

func (m *Manager) routerInitializer(router *routerpkg.Router, netConnection *netadapter.NetConnection) {
	// isStopping flag is raised the moment that the connection associated with this router is disconnected
	// errChan is used by the Flow goroutines to return to runFlows when an error occurs.
	// They are both initialized here and passed to register flows.
	isStopping := uint32(0)
	errChan := make(chan error)

	receiveVersionRoute, sendVersionRoute := registerHandshakeRoutes(router)

	// After flows were registered - spawn a new thread that will wait for connection to finish initializing
	// and start receiving messages
	spawn("routerInitializer-runFlows", func() {
		m.routersWaitGroup.Add(1)
		defer m.routersWaitGroup.Done()

		if atomic.LoadUint32(&m.isClosed) == 1 {
			panic(errors.Errorf("tried to initialize router when the protocol manager is closed"))
		}

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

		var flows []*common.Flow
		if peer.ProtocolVersion() == 3 {
			flows = v3.Register(m, router, errChan, &isStopping)
		}

		removeHandshakeRoutes(router)

		flowsWaitGroup := &sync.WaitGroup{}
		err = m.runFlows(flows, peer, errChan, flowsWaitGroup)
		if err != nil {
			m.handleError(err, netConnection, router.OutgoingRoute())
			// We call `flowsWaitGroup.Wait()` in two places instead of deferring, because
			// we already defer `m.routersWaitGroup.Done()`, so we try to avoid error prone
			// and confusing use of multiple dependent defers.
			flowsWaitGroup.Wait()
			return
		}
		flowsWaitGroup.Wait()
	})
}

func (m *Manager) handleError(err error, netConnection *netadapter.NetConnection, outgoingRoute *routerpkg.Route) {
	if protocolErr := (protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
		if m.context.Config().EnableBanning && protocolErr.ShouldBan {
			log.Warnf("Banning %s (reason: %s)", netConnection, protocolErr.Cause)

			err := m.context.ConnectionManager().Ban(netConnection)
			if err != nil && !errors.Is(err, connmanager.ErrCannotBanPermanent) {
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

// RegisterFlow registers a flow to the given router.
func (m *Manager) RegisterFlow(name string, router *routerpkg.Router, messageTypes []appmessage.MessageCommand, isStopping *uint32,
	errChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow {

	route, err := router.AddIncomingRoute(name, messageTypes)
	if err != nil {
		panic(err)
	}

	return m.registerFlowForRoute(route, name, isStopping, errChan, initializeFunc)
}

// RegisterFlowWithCapacity registers a flow to the given router with a custom capacity.
func (m *Manager) RegisterFlowWithCapacity(name string, capacity int, router *routerpkg.Router,
	messageTypes []appmessage.MessageCommand, isStopping *uint32,
	errChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow {

	route, err := router.AddIncomingRouteWithCapacity(name, capacity, messageTypes)
	if err != nil {
		panic(err)
	}

	return m.registerFlowForRoute(route, name, isStopping, errChan, initializeFunc)
}

func (m *Manager) registerFlowForRoute(route *routerpkg.Route, name string, isStopping *uint32,
	errChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow {

	return &common.Flow{
		Name: name,
		ExecuteFunc: func(peer *peerpkg.Peer) {
			err := initializeFunc(route, peer)
			if err != nil {
				m.context.HandleError(err, name, isStopping, errChan)
				return
			}
		},
	}
}

// RegisterOneTimeFlow registers a one-time flow (that exits once some operations are done) to the given router.
func (m *Manager) RegisterOneTimeFlow(name string, router *routerpkg.Router, messageTypes []appmessage.MessageCommand,
	isStopping *uint32, stopChan chan error, initializeFunc common.FlowInitializeFunc) *common.Flow {

	route, err := router.AddIncomingRoute(name, messageTypes)
	if err != nil {
		panic(err)
	}

	return &common.Flow{
		Name: name,
		ExecuteFunc: func(peer *peerpkg.Peer) {
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
	receiveVersionRoute, err := router.AddIncomingRoute("recieveVersion - incoming", []appmessage.MessageCommand{appmessage.CmdVersion})
	if err != nil {
		panic(err)
	}

	sendVersionRoute, err = router.AddIncomingRoute("sendVersion - incoming", []appmessage.MessageCommand{appmessage.CmdVerAck})
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
