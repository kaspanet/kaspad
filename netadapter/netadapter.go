package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/id"
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
type RouterInitializer func() (*Router, error)

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
	idsToRouters     map[*id.ID]*Router
	sync.RWMutex
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningAddrs []string) (*NetAdapter, error) {
	netAdapterID, err := id.GenerateID()
	if err != nil {
		return nil, err
	}
	s, err := grpcserver.NewGRPCServer(listeningAddrs)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		id:     netAdapterID,
		server: s,

		connectionIDs:    make(map[server.Connection]*id.ID),
		idsToConnections: make(map[*id.ID]server.Connection),
		idsToRouters:     make(map[*id.ID]*Router),
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
		connection.SetOnDisconnectedHandler(func() error {
			na.unregisterConnection(connection)
			return router.Close()
		})
		router.SetOnIDReceivedHandler(func(id *id.ID) {
			na.registerConnection(connection, router, id)
		})

		spawn(func() { na.startReceiveLoop(connection, router) })
		spawn(func() { na.startSendLoop(connection, router) })
		return nil
	}
}

func (na *NetAdapter) registerConnection(connection server.Connection, router *Router, id *id.ID) {
	na.server.AddConnection(connection)

	na.connectionIDs[connection] = id
	na.idsToConnections[id] = connection
	na.idsToRouters[id] = router
}

func (na *NetAdapter) unregisterConnection(connection server.Connection) {
	na.server.RemoveConnection(connection)

	id, ok := na.connectionIDs[connection]
	if !ok {
		return
	}

	delete(na.connectionIDs, connection)
	delete(na.idsToConnections, id)
	delete(na.idsToRouters, id)
}

func (na *NetAdapter) startReceiveLoop(connection server.Connection, router *Router) {
	for atomic.LoadUint32(&na.stop) == 0 {
		message, err := connection.Receive()
		if err != nil {
			log.Warnf("Failed to receive from %s: %s", connection, err)
			break
		}
		err = router.RouteIncomingMessage(message)
		if err != nil {
			// TODO(libp2p): This should never happen, do something more severe
			log.Warnf("Failed to route input message from %s: %s", connection, err)
			break
		}
	}

	if connection.IsConnected() {
		err := connection.Disconnect()
		if err != nil {
			log.Warnf("Failed to disconnect from %s: %s", connection, err)
		}
	}
}

func (na *NetAdapter) startSendLoop(connection server.Connection, router *Router) {
	for atomic.LoadUint32(&na.stop) == 0 {
		message := router.ReadOutgoingMessage()
		err := connection.Send(message)
		if err != nil {
			log.Warnf("Failed to send to %s: %s", connection, err)
			break
		}
	}

	if connection.IsConnected() {
		err := connection.Disconnect()
		if err != nil {
			log.Warnf("Failed to disconnect from %s: %s", connection, err)
		}
	}
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
func (na *NetAdapter) Broadcast(ids []*id.ID, message wire.Message) {
	na.RLock()
	defer na.RUnlock()
	for _, id := range ids {
		router, ok := na.idsToRouters[id]
		if !ok {
			log.Warnf("id %s is not registered", id)
			continue
		}
		router.WriteOutgoingMessage(message)
	}
}

// GetBestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (na *NetAdapter) GetBestLocalAddress() *wire.NetAddress {
	//TODO(libp2p) Implement this, and check reachability to the other node
	panic("unimplemented")
}
