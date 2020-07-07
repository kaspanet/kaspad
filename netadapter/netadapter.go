package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpc"
)

// NetAdapter is an adapter to the net
type NetAdapter struct {
	server            server.Server
	routerInitializer func(peer *Peer) *Router
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningPort string) (*NetAdapter, error) {
	server, err := grpc.NewGRPCServer(listeningPort)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		server: server,
	}

	newConnectionHandler := adapter.buildNewConnectionHandler()
	server.SetNewConnectionHandler(newConnectionHandler)

	return &adapter, nil
}

func (na *NetAdapter) buildNewConnectionHandler() func(connection server.Connection) {
	return func(connection server.Connection) {
		peer := NewPeer(connection)
		router := na.routerInitializer(peer)

		for {
			message, err := peer.connection.Receive()
			if err != nil {
				// TODO: properly handle error
				panic(err)
			}
			router.RouteMessage(message)
		}
	}
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer func(peer *Peer) *Router) {
	na.routerInitializer = routerInitializer
}

func (na *NetAdapter) Close() error {
	return na.server.Close()
}
