package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_RequestSelectedTip) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestSelectedTip{}, nil
}

func (x *KaspadMessage_RequestSelectedTip) fromAppMessage(_ *appmessage.MsgRequestSelectedTip) error {
	return nil
}
