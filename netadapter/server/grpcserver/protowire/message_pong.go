package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_Pong) toWireMessage() (*wire.MsgPing, error) {
	return &wire.MsgPing{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromWireMessage(msgPong *wire.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
