package netadapter

import (
	"fmt"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/netadapter/server"
)

// NetConnection is a wrapper to a server connection for use by services external to NetAdapter
type NetConnection struct {
	connection server.Connection
	id         *id.ID
}

func newNetConnection(connection server.Connection, id *id.ID) *NetConnection {
	return &NetConnection{
		connection: connection,
		id:         id,
	}
}

func (c *NetConnection) String() string {
	return fmt.Sprintf("<%s: %s>", c.id, c.connection)
}

// ID returns the ID associated with this connection
func (c *NetConnection) ID() *id.ID {
	return c.id
}

// Address returns the address associated with this connection
func (c *NetConnection) Address() string {
	return c.connection.Address().String()
}

// SetOnBadMessageHandler sets a handler function
// for bad messages
func (c *NetConnection) SetOnBadMessageHandler(onBadMessageHandler server.OnBadMessageHandler) {
	c.connection.SetOnBadMessageHandler(onBadMessageHandler)
}
