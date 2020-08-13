package protowire

import (
	"github.com/kaspanet/kaspad/network/domainmessage"
)

func (x *KaspadMessage_Pong) toDomainMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgPong{
		Nonce: x.Pong.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromDomainMessage(msgPong *domainmessage.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
