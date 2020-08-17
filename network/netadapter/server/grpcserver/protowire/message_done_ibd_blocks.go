package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_DoneIBDBlocks) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneIBDBlocks) fromAppMessage(_ *appmessage.MsgDoneIBDBlocks) error {
	return nil
}
