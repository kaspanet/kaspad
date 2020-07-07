package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/wire"
)

type Peer struct {
	connection server.Connection
	router     *Router
}

func (p *Peer) SendMessage(message wire.Message) error {
	return p.connection.Send(message)
}
