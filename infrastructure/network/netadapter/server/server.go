package server

import (
	"fmt"
	"net"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// OnConnectedHandler is a function that is to be called
// once a new Connection is successfully established.
type OnConnectedHandler func(connection Connection) error

// OnDisconnectedHandler is a function that is to be
// called once a Connection has been disconnected.
type OnDisconnectedHandler func()

// OnInvalidMessageHandler is a function that is to be called when
// an invalid message (cannot be parsed/doesn't have a route)
// was received from a connection.
type OnInvalidMessageHandler func(err error)

// Server represents a server.
type Server interface {
	Start() error
	Stop() error
	SetOnConnectedHandler(onConnectedHandler OnConnectedHandler)
}

// P2PServer represents a p2p server.
type P2PServer interface {
	Server
	Connect(address string) (Connection, error)
}

// Connection represents a server connection.
type Connection interface {
	fmt.Stringer
	Start(router *router.Router)
	Disconnect()
	IsConnected() bool
	IsOutbound() bool
	SetOnDisconnectedHandler(onDisconnectedHandler OnDisconnectedHandler)
	SetOnInvalidMessageHandler(onInvalidMessageHandler OnInvalidMessageHandler)
	Address() *net.TCPAddr
}
