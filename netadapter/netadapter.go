package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
)

// RouterInitializer is a function that initializes a new
// router to be used with a new connection
type RouterInitializer func(connection *Connection) (*Router, error)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "routes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	server            server.Server
	routerInitializer RouterInitializer
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningAddrs []string) (*NetAdapter, error) {
	s, err := grpcserver.NewGRPCServer(listeningAddrs)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		server: s,
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
	return na.server.Stop()
}

func (na *NetAdapter) newOnConnectedHandler() server.OnConnectedHandler {
	return func(serverConnection server.Connection) {
		connection := NewConnection(serverConnection)
		router, err := na.routerInitializer(connection)
		if err != nil {
			// TODO(libp2p): properly handle error
			panic(err)
		}
		serverConnection.SetOnDisconnectedHandler(func() {
			err := router.Close()
			if err != nil {
				// TODO(libp2p): properly handle error
				panic(err)
			}
		})

		for {
			message, err := connection.connection.Receive()
			if err != nil {
				// TODO(libp2p): properly handle error
				panic(err)
			}
			router.RouteMessage(message)
		}
	}
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer RouterInitializer) {
	na.routerInitializer = routerInitializer
}
