package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_IbdBlockLocatorHighestHashNotFound) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgIBDBlockLocatorHighestHashNotFound{}, nil
}

func (x *KaspadMessage_IbdBlockLocatorHighestHashNotFound) fromAppMessage(message *appmessage.MsgIBDBlockLocatorHighestHashNotFound) error {
	x.IbdBlockLocatorHighestHashNotFound = &IbdBlockLocatorHighestHashNotFoundMessage{}
	return nil
}
