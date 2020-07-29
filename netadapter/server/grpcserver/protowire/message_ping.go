package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_Ping) toWireMessage() (wire.Message, error) {
	return &wire.MsgPing{
		Nonce: x.Ping.Nonce,
	}, nil
}

func (x *KaspadMessage_Ping) fromWireMessage(msgPing *wire.MsgPing) error {
	x.Ping = &PingMessage{
		Nonce: msgPing.Nonce,
	}
	return nil
}
