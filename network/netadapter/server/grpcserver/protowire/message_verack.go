package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_Verack) toDomainMessage() (appmessage.Message, error) {
	return &appmessage.MsgVerAck{}, nil
}

func (x *KaspadMessage_Verack) fromDomainMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
