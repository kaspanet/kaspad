package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_ResolveFinalityConflictRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.ResolveFinalityConflictRequestMessage{
		FinalityBlockHash: x.ResolveFinalityConflictRequest.FinalityBlockHash,
	}, nil
}

func (x *KaspadMessage_ResolveFinalityConflictRequest) fromAppMessage(message *appmessage.ResolveFinalityConflictRequestMessage) error {
	x.ResolveFinalityConflictRequest = &ResolveFinalityConflictRequestMessage{
		FinalityBlockHash: message.FinalityBlockHash,
	}
	return nil
}

func (x *KaspadMessage_ResolveFinalityConflictResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.ResolveFinalityConflictResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.ResolveFinalityConflictResponse.Error.Message}
	}
	return &appmessage.ResolveFinalityConflictResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_ResolveFinalityConflictResponse) fromAppMessage(message *appmessage.ResolveFinalityConflictResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ResolveFinalityConflictResponse = &ResolveFinalityConflictResponseMessage{
		Error: err,
	}
	return nil
}
