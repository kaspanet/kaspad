package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestNextIBDBlocks) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestNextIBDBlocks{}, nil
}

func (x *KaspadMessage_RequestNextIBDBlocks) fromAppMessage(_ *appmessage.MsgRequestNextIBDBlocks) error {
	return nil
}
