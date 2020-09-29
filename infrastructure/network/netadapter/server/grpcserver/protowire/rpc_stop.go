package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_StopRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.StopRequestMessage{}, nil
}

func (x *KaspadMessage_StopRequest) fromAppMessage(_ *appmessage.StopRequestMessage) error {
	x.StopRequest = &StopRequest{}
	return nil
}

func (x *KaspadMessage_StopResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.StopResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.StopResponse.Error.Message}
	}
	return &appmessage.StopResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_StopResponse) fromAppMessage(message *appmessage.StopResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StopResponse = &StopResponse{
		Error: err,
	}
	return nil
}
