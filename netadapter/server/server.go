package server

import (
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
}

// Connection represents a p2p server connection.
type Connection interface {
	Send(message wire.Message) error
	Receive() (wire.Message, error)
	Disconnect() error
	IsConnected() bool
	SetOnDisconnectedHandler(onDisconnectedHandler OnDisconnectedHandler)
}
