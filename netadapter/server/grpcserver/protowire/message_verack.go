package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_Verack) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgVerAck{}, nil
}

func (x *KaspadMessage_Verack) fromWireMessage(_ *domainmessage.MsgVerAck) error {
	return nil
}
