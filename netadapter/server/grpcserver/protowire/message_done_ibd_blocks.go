package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_DoneIBDBlocks) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneIBDBlocks) fromWireMessage(_ *domainmessage.MsgDoneIBDBlocks) error {
	return nil
}
