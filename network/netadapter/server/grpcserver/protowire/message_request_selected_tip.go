package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_RequestSelectedTip) toDomainMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestSelectedTip{}, nil
}

func (x *KaspadMessage_RequestSelectedTip) fromDomainMessage(_ *appmessage.MsgRequestSelectedTip) error {
	return nil
}
