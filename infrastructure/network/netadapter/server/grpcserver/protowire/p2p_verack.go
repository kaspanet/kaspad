package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_Verack) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgVerAck{}, nil
}

func (x *KaspadMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
