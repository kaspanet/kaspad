package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestPruningPointHash) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestPruningPointHashMessage{}, nil
}

func (x *KaspadMessage_RequestPruningPointHash) fromAppMessage(_ *appmessage.MsgRequestPruningPointHashMessage) error {
	return nil
}
