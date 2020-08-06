package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_RequestSelectedTip) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgRequestSelectedTip{}, nil
}

func (x *KaspadMessage_RequestSelectedTip) fromWireMessage(_ *domainmessage.MsgRequestSelectedTip) error {
	return nil
}
