package server

import "github.com/kaspanet/kaspad/wire"

// Server represents a p2p server.
type Server interface {
	Connect(address string) (Connection, error)
	Connections() []Connection
	Close() error
}

// Connection represents a p2p server connection.
type Connection interface {
	Send(message wire.Message) error
	Receive() (wire.Message, error)
	Disconnect() error
}
