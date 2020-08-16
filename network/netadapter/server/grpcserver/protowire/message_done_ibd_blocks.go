package protowire

import "github.com/kaspanet/kaspad/network/domainmessage"

func (x *KaspadMessage_DoneIBDBlocks) toDomainMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneIBDBlocks) fromDomainMessage(_ *domainmessage.MsgDoneIBDBlocks) error {
	return nil
}
