package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_Pong) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgPong{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromAppMessage(msgPong *appmessage.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
