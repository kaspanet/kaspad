package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
)

func (x *KaspadMessage_Pong) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgPong{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromWireMessage(msgPong *domainmessage.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
