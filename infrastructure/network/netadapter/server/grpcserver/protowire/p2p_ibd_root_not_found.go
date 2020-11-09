package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_IbdRootNotFound) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgIBDRootNotFound{}, nil
}

func (x *KaspadMessage_IbdRootNotFound) fromAppMessage(_ *appmessage.MsgIBDRootNotFound) error {
	return nil
}
