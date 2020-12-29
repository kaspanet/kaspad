package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestIBDRootHash) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestIBDRootHash{}, nil
}

func (x *KaspadMessage_RequestIBDRootHash) fromAppMessage(_ *appmessage.MsgRequestIBDRootHash) error {
	return nil
}
