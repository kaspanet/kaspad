package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_SubmitBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.SubmitBlockRequestMessage{}, nil
}

func (x *KaspadMessage_SubmitBlockRequest) fromAppMessage(_ *appmessage.SubmitBlockRequestMessage) error {
	return nil
}
