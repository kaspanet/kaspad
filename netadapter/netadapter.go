package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/id"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// RouterInitializer is a function that initializes a new
// router to be used with a new connection
type RouterInitializer func() (*routerpkg.Router, error)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "routes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	id                *id.ID
	server            server.Server
	routerInitializer RouterInitializer
	stop              uint32

	connectionIDs    map[server.Connection]*id.ID
	idsToConnections map[*id.ID]server.Connection
	idsToRouters     map[*id.ID]*routerpkg.Router
	sync.RWMutex
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningAddrs []string) (*NetAdapter, error) {
	netAdapterId, err := id.GenerateID()
	if err != nil {
		return nil, err
	}
	s, err := grpcserver.NewGRPCServer(listeningAddrs)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		id:     netAdapterId,
		server: s,

		connectionIDs:    make(map[server.Connection]*id.ID),
		idsToConnections: make(map[*id.ID]server.Connection),
		idsToRouters:     make(map[*id.ID]*routerpkg.Router),
	}

	onConnectedHandler := adapter.newOnConnectedHandler()
	adapter.server.SetOnConnectedHandler(onConnectedHandler)

	return &adapter, nil
}

// Start begins the operation of the NetAdapter
func (na *NetAdapter) Start() error {
	err := na.server.Start()
	if err != nil {
		return err
	}

	// TODO(libp2p): Replace with real connection manager
	cfg := config.ActiveConfig()
	for _, connectPeer := range cfg.ConnectPeers {
		_, err := na.server.Connect(connectPeer)
		if err != nil {
			log.Errorf("Error connecting to %s: %+v", connectPeer, err)
		}
	}

	return nil
}

// Stop safely closes the NetAdapter
func (na *NetAdapter) Stop() error {
	if atomic.AddUint32(&na.stop, 1) != 1 {
		return errors.New("net adapter stopped more than once")
	}
	return na.server.Stop()
}

func (na *NetAdapter) newOnConnectedHandler() server.OnConnectedHandler {
	return func(connection server.Connection) error {
		router, err := na.routerInitializer()
		if err != nil {
			return err
		}
		connection.SetRouter(router)

		router.SetOnRouteCapacityReachedHandler(func() {
			err := connection.Disconnect()
			if err != nil {
				log.Warnf("Failed to disconnect from %s", connection)
			}
		})
		connection.SetOnDisconnectedHandler(func() error {
			na.unregisterConnection(connection)
			return router.Close()
		})
		router.SetOnIDReceivedHandler(func(id *id.ID) {
			na.registerConnection(connection, router, id)
		})
		return nil
	}
}

func (na *NetAdapter) registerConnection(connection server.Connection, router *routerpkg.Router, id *id.ID) {
	na.server.AddConnection(connection)

	na.connectionIDs[connection] = id
	na.idsToConnections[id] = connection
	na.idsToRouters[id] = router
}

func (na *NetAdapter) unregisterConnection(connection server.Connection) {
	na.server.RemoveConnection(connection)

	connectionID, ok := na.connectionIDs[connection]
	if !ok {
		return
	}

	delete(na.connectionIDs, connection)
	delete(na.idsToConnections, connectionID)
	delete(na.idsToRouters, connectionID)
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer RouterInitializer) {
	na.routerInitializer = routerInitializer
}

// ID returns this netAdapter's ID in the network
func (na *NetAdapter) ID() *id.ID {
	return na.id
}

// Broadcast sends the given `message` to every peer corresponding
// to each ID in `ids`
func (na *NetAdapter) Broadcast(connectionIDs []*id.ID, message wire.Message) error {
	na.RLock()
	defer na.RUnlock()
	for _, connectionID := range connectionIDs {
		router, ok := na.idsToRouters[connectionID]
		if !ok {
			log.Warnf("connectionID %s is not registered", connectionID)
			continue
		}
		route, err := router.IncomingRoute(message)
		if err != nil {
			return err
		}
		err = route.Enqueue(message)
		if err != nil {
			return err
		}
	}
	return nil
}
