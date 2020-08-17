package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_DoneIBDBlocks) toDomainMessage() (appmessage.Message, error) {
	return &appmessage.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneIBDBlocks) fromDomainMessage(_ *appmessage.MsgDoneIBDBlocks) error {
	return nil
}
