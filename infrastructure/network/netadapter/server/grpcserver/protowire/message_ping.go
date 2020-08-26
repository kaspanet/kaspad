package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_Ping) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgPing{
		Nonce: x.Ping.Nonce,
	}, nil
}

func (x *KaspadMessage_Ping) fromAppMessage(msgPing *appmessage.MsgPing) error {
	x.Ping = &PingMessage{
		Nonce: msgPing.Nonce,
	}
	return nil
}
