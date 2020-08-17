package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_RequestNextIBDBlocks) toDomainMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestNextIBDBlocks{}, nil
}

func (x *KaspadMessage_RequestNextIBDBlocks) fromDomainMessage(_ *appmessage.MsgRequestNextIBDBlocks) error {
	return nil
}
