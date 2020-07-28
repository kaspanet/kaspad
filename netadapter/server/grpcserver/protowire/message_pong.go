package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_Pong) toWireMessage() (*wire.MsgPing, error) {
	return &wire.MsgPing{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromWireMessage(msgPing *wire.MsgPing) {
	x.Pong = &PongMessage{
		Nonce: msgPing.Nonce,
	}
}
