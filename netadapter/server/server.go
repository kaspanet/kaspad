package server

import (
	"github.com/kaspanet/kaspad/wire"
)

// PeerConnectedHandler is a function that is to be called
// once a new Connection is successfully established.
type PeerConnectedHandler func(connection Connection)

// PeerDisconnectedHandler is a function that is to be
// called once a Connection has been disconnected.
type PeerDisconnectedHandler func()

// Server represents a p2p server.
type Server interface {
	Connect(address string) (Connection, error)
	Connections() []Connection
	Start() error
	Stop() error
	SetPeerConnectedHandler(peerConnectedHandler PeerConnectedHandler)
}

// Connection represents a p2p server connection.
type Connection interface {
	Send(message wire.Message) error
	Receive() (wire.Message, error)
	Disconnect() error
	SetPeerDisconnectedHandler(peerDisconnectedHandler PeerDisconnectedHandler)
}
