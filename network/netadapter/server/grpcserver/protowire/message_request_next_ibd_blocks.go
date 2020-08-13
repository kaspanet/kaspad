package protowire

import "github.com/kaspanet/kaspad/network/domainmessage"

func (x *KaspadMessage_RequestNextIBDBlocks) toDomainMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgRequestNextIBDBlocks{}, nil
}

func (x *KaspadMessage_RequestNextIBDBlocks) fromDomainMessage(_ *domainmessage.MsgRequestNextIBDBlocks) error {
	return nil
}
