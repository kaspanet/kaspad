package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_DoneHeaders) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgDoneHeaders{}, nil
}

func (x *KaspadMessage_DoneHeaders) fromAppMessage(_ *appmessage.MsgDoneHeaders) error {
	return nil
}
