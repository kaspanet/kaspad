package protowire

import "github.com/kaspanet/kaspad/network/domainmessage"

func (x *KaspadMessage_Verack) toDomainMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgVerAck{}, nil
}

func (x *KaspadMessage_Verack) fromDomainMessage(_ *domainmessage.MsgVerAck) error {
	return nil
}
