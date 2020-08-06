package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_RequestNextIBDBlocks) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgRequestNextIBDBlocks{}, nil
}

func (x *KaspadMessage_RequestNextIBDBlocks) fromWireMessage(_ *domainmessage.MsgRequestNextIBDBlocks) error {
	return nil
}
