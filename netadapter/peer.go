package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/wire"
)

// Peer represents a remote peer in a network
type Peer struct {
	connection server.Connection
}

// NewPeer creates a new Peer wrapping the given connection
func NewPeer(connection server.Connection) *Peer {
	return &Peer{
		connection: connection,
	}
}

// SendMessage sends the given message to the remote peer
func (p *Peer) SendMessage(message wire.Message) error {
	return p.connection.Send(message)
}
