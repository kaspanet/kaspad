package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
)

func (x *KaspadMessage_Ping) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgPing{
		Nonce: x.Ping.Nonce,
	}, nil
}

func (x *KaspadMessage_Ping) fromWireMessage(msgPing *domainmessage.MsgPing) error {
	x.Ping = &PingMessage{
		Nonce: msgPing.Nonce,
	}
	return nil
}
