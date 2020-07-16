package netadapter

import (
	"fmt"

	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/netadapter/server"
)

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

func (c *NetConnection) ID() *id.ID {
	return c.id
}
