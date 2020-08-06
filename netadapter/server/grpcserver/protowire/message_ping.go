package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
)

func (x *KaspadMessage_Ping) toDomainMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgPing{
		Nonce: x.Ping.Nonce,
	}, nil
}

func (x *KaspadMessage_Ping) fromDomainMessage(msgPing *domainmessage.MsgPing) error {
	x.Ping = &PingMessage{
		Nonce: msgPing.Nonce,
	}
	return nil
}
