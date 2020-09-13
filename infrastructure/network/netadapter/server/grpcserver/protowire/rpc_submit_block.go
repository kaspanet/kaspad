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
	var err *appmessage.RPCError
	if x.SubmitBlockResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.SubmitBlockResponse.Error.Message}
	}
	return &appmessage.SubmitBlockResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_SubmitBlockResponse) fromAppMessage(message *appmessage.SubmitBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.SubmitBlockResponse = &SubmitBlockResponseMessage{
		Error: err,
	}
	return nil
}
