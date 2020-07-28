package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_Ping) toWireMessage() (*wire.MsgPing, error) {
	return &wire.MsgPing{
		Nonce: x.Ping.Nonce,
	}, nil
}

func (x *KaspadMessage_Ping) fromWireMessage(msgPing *wire.MsgPing) {
	x.Ping = &PingMessage{
		Nonce: msgPing.Nonce,
	}
}
