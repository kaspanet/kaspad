package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestNextHeaders) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestNextHeaders{}, nil
}

func (x *KaspadMessage_RequestNextHeaders) fromAppMessage(_ *appmessage.MsgRequestNextHeaders) error {
	return nil
}
