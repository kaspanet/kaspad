package protowire

import "github.com/kaspanet/kaspad/network/domainmessage"

func (x *KaspadMessage_RequestSelectedTip) toDomainMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgRequestSelectedTip{}, nil
}

func (x *KaspadMessage_RequestSelectedTip) fromDomainMessage(_ *domainmessage.MsgRequestSelectedTip) error {
	return nil
}
