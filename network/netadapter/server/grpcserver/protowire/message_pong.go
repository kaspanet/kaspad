package protowire

import (
	"github.com/kaspanet/kaspad/network/appmessage"
)

func (x *KaspadMessage_Pong) toDomainMessage() (appmessage.Message, error) {
	return &appmessage.MsgPong{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromDomainMessage(msgPong *appmessage.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
