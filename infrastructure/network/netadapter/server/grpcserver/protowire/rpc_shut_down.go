package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_ShutDownRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.ShutDownRequestMessage{}, nil
}

func (x *KaspadMessage_ShutDownRequest) fromAppMessage(_ *appmessage.ShutDownRequestMessage) error {
	x.ShutDownRequest = &ShutDownRequestMessage{}
	return nil
}

func (x *KaspadMessage_ShutDownResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.ShutDownResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.ShutDownResponse.Error.Message}
	}
	return &appmessage.ShutDownResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_ShutDownResponse) fromAppMessage(message *appmessage.ShutDownResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ShutDownResponse = &ShutDownResponseMessage{
		Error: err,
	}
	return nil
}
