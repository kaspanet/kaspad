package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_ConnectToPeerRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.ConnectToPeerRequestMessage{
		Address:     x.ConnectToPeerRequest.Address,
		IsPermanent: x.ConnectToPeerRequest.IsPermanent,
	}, nil
}

func (x *KaspadMessage_ConnectToPeerRequest) fromAppMessage(message *appmessage.ConnectToPeerRequestMessage) error {
	x.ConnectToPeerRequest = &ConnectToPeerRequestMessage{
		Address:     message.Address,
		IsPermanent: message.IsPermanent,
	}
	return nil
}

func (x *KaspadMessage_ConnectToPeerResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.ConnectToPeerResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.ConnectToPeerResponse.Error.Message}
	}
	return &appmessage.ConnectToPeerResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_ConnectToPeerResponse) fromAppMessage(message *appmessage.ConnectToPeerResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ConnectToPeerResponse = &ConnectToPeerResponseMessage{
		Error: err,
	}
	return nil
}
