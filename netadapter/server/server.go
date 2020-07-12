package server

import (
	"fmt"
	"net"

	"github.com/kaspanet/kaspad/wire"
)

// OnConnectedHandler is a function that is to be called
// once a new Connection is successfully established.
type OnConnectedHandler func(connection Connection) error

// OnDisconnectedHandler is a function that is to be
// called once a Connection has been disconnected.
type OnDisconnectedHandler func() error

// Server represents a p2p server.
type Server interface {
	Connect(address string) (Connection, error)
	Connections() []Connection
	Start() error
	Stop() error
	SetOnConnectedHandler(onConnectedHandler OnConnectedHandler)
	// TODO(libp2p): Move AddConnection and RemoveConnection to connection manager
	AddConnection(connection Connection) error
	RemoveConnection(connection Connection) error
}

// Connection represents a p2p server connection.
type Connection interface {
	fmt.Stringer
	Send(message wire.Message) error
	Receive() (wire.Message, error)
	Disconnect() error
	IsConnected() bool
	SetOnDisconnectedHandler(onDisconnectedHandler OnDisconnectedHandler)
	Address() net.Addr
}
