package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
	"sync/atomic"
)

// RouterInitializer is a function that initializes a new
// router to be used with a new connection
type RouterInitializer func() (*Router, error)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "inputRoutes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	id                *ID
	server            server.Server
	routerInitializer RouterInitializer
	stop              int32

	connectionIDs    map[server.Connection]*ID
	idsToConnections map[*ID]server.Connection
	idsToRouters     map[*ID]*Router
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningAddrs []string) (*NetAdapter, error) {
	id, err := GenerateID()
	if err != nil {
		return nil, err
	}
	s, err := grpcserver.NewGRPCServer(listeningAddrs)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		id:     id,
		server: s,

		connectionIDs:    make(map[server.Connection]*ID),
		idsToConnections: make(map[*ID]server.Connection),
		idsToRouters:     make(map[*ID]*Router),
	}

	onConnectedHandler := adapter.newOnConnectedHandler()
	adapter.server.SetOnConnectedHandler(onConnectedHandler)

	return &adapter, nil
}

// Start begins the operation of the NetAdapter
func (na *NetAdapter) Start() error {
	return na.server.Start()
}

// Stop safely closes the NetAdapter
func (na *NetAdapter) Stop() error {
	if atomic.AddInt32(&na.stop, 1) != 1 {
		log.Warnf("Net adapter stopped more than once")
		return nil
	}
	return na.server.Stop()
}

func (na *NetAdapter) newOnConnectedHandler() server.OnConnectedHandler {
	return func(connection server.Connection) error {
		router, err := na.routerInitializer()
		if err != nil {
			return err
		}
		connection.SetOnDisconnectedHandler(func() error {
			na.unregisterConnection(connection)
			return router.Close()
		})
		router.SetOnIDReceivedHandler(func(id *ID) {
			na.registerConnection(connection, router, id)
		})

		na.startReceiveLoop(connection, router)
		na.startSendLoop(connection, router)
		return nil
	}
}

func (na *NetAdapter) registerConnection(connection server.Connection, router *Router, id *ID) {
	na.connectionIDs[connection] = id
	na.idsToConnections[id] = connection
	na.idsToRouters[id] = router
}

func (na *NetAdapter) unregisterConnection(connection server.Connection) {
	id, ok := na.connectionIDs[connection]
	if !ok {
		return
	}

	delete(na.connectionIDs, connection)
	delete(na.idsToConnections, id)
	delete(na.idsToRouters, id)
}

func (na *NetAdapter) startReceiveLoop(connection server.Connection, router *Router) {
	spawn(func() {
		for {
			if atomic.LoadInt32(&na.stop) != 0 {
				err := connection.Disconnect()
				if err != nil {
					log.Warnf("Failed to disconnect from %s: %s", connection, err)
				}
				return
			}

			message, err := connection.Receive()
			if err != nil {
				log.Warnf("Failed to receive from %s: %s", connection, err)
				err := connection.Disconnect()
				if err != nil {
					log.Warnf("Failed to disconnect from %s: %s", connection, err)
				}
			}
			router.RouteInputMessage(message)
		}
	})
}

func (na *NetAdapter) startSendLoop(connection server.Connection, router *Router) {
	spawn(func() {
		for {
			if atomic.LoadInt32(&na.stop) != 0 {
				err := connection.Disconnect()
				if err != nil {
					log.Warnf("Failed to disconnect from %s: %s", connection, err)
				}
				return
			}

			message := router.TakeOutputMessage()
			err := connection.Send(message)
			if err != nil {
				log.Warnf("Failed to send to %s: %s", connection, err)
				err := connection.Disconnect()
				if err != nil {
					log.Warnf("Failed to disconnect from %s: %s", connection, err)
				}
			}
		}
	})
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer RouterInitializer) {
	na.routerInitializer = routerInitializer
}

// ID returns this netAdapter's ID in the network
func (na *NetAdapter) ID() *ID {
	return na.id
}
