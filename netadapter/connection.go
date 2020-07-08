package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/wire"
)

// Connection represents a connection to remote network peer
type Connection struct {
	connection server.Connection
}

// NewConnection creates a new Connection wrapping the given
// server.Connection
func NewConnection(connection server.Connection) *Connection {
	return &Connection{
		connection: connection,
	}
}

// SendMessage sends the given message to the remote peer
func (p *Connection) SendMessage(message wire.Message) error {
	return p.connection.Send(message)
}

// Disconnect disconnects the connection
func (p *Connection) Disconnect() error {
	return p.connection.Disconnect()
}
