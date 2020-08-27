package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_SubmitBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.SubmitBlockRequestMessage{
		BlockHex: x.SubmitBlockRequest.BlockHex,
	}, nil
}

func (x *KaspadMessage_SubmitBlockRequest) fromAppMessage(message *appmessage.SubmitBlockRequestMessage) error {
	x.SubmitBlockRequest = &SubmitBlockRequestMessage{
		BlockHex: message.BlockHex,
	}
	return nil
}

func (x *KaspadMessage_SubmitBlockResponse) toAppMessage() (appmessage.Message, error) {
	return &appmessage.SubmitBlockRequestMessage{}, nil
}

func (x *KaspadMessage_SubmitBlockResponse) fromAppMessage(_ *appmessage.SubmitBlockResponseMessage) error {
	return nil
}
