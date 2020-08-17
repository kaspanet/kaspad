package protowire

import (
	"github.com/kaspanet/kaspad/network/appmessage"
)

func (x *KaspadMessage_Ping) toDomainMessage() (appmessage.Message, error) {
	return &appmessage.MsgPing{
		Nonce: x.Ping.Nonce,
	}, nil
}

func (x *KaspadMessage_Ping) fromDomainMessage(msgPing *appmessage.MsgPing) error {
	x.Ping = &PingMessage{
		Nonce: msgPing.Nonce,
	}
	return nil
}
