package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
)

// RouterInitializer is a function that initializes a new
// router to be used with a newly connected peer
type RouterInitializer func(peer *Peer) (*Router, error)

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
func NewNetAdapter(listeningPort string) (*NetAdapter, error) {
	server, err := grpcserver.NewGRPCServer(listeningPort)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		server: server,
	}

	peerConnectedHandler := adapter.newPeerConnectedHandler()
	server.SetPeerConnectedHandler(peerConnectedHandler)

	return &adapter, nil
}

func (na *NetAdapter) newPeerConnectedHandler() server.PeerConnectedHandler {
	return func(connection server.Connection) {
		peer := NewPeer(connection)
		router, err := na.routerInitializer(peer)
		if err != nil {
			// TODO(libp2p): properly handle error
			panic(err)
		}

		for {
			message, err := peer.connection.Receive()
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

// Close safely closes the netAdapter
func (na *NetAdapter) Close() error {
	return na.server.Close()
}
