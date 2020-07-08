package server

import (
	"github.com/kaspanet/kaspad/wire"
)

// PeerConnectedHandler is a function that is to be called
// once a new Connection is successfully established.
type PeerConnectedHandler func(connection Connection)

// Server represents a p2p server.
type Server interface {
	SetPeerConnectedHandler(peerConnectedHandler PeerConnectedHandler)
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
