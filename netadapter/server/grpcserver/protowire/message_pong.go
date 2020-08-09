package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_Pong) toWireMessage() (wire.Message, error) {
	return &wire.MsgPong{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromWireMessage(msgPong *wire.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
