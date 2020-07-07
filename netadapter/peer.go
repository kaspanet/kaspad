package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/wire"
)

type Peer struct {
	connection server.Connection
}

func NewPeer(connection server.Connection) *Peer {
	return &Peer{
		connection: connection,
	}
}

func (p *Peer) SendMessage(message wire.Message) error {
	return p.connection.Send(message)
}
