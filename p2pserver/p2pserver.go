package p2pserver

import "github.com/kaspanet/kaspad/wire"

// Server represents a p2p server.
type Server interface {
	BroadcastExceptConnection(msg wire.Message, exceptConnection Connection)
	Connect(address string) (Connection, error)
	Connections() []Connection
}

// Connection represents a p2p server connection.
type Connection interface {
	Send(message wire.Message) error
	Receive() (wire.Message, error)
	Disconnect() error
	AddBanScore(persistent, transient uint32, reason string) (banned bool)
	IsInbound() bool
	String() string
}
