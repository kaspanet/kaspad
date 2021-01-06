package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetCurrentNetworkRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetCurrentNetworkRequestMessage{}, nil
}

func (x *KaspadMessage_GetCurrentNetworkRequest) fromAppMessage(_ *appmessage.GetCurrentNetworkRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetCurrentNetworkResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetCurrentNetworkResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetCurrentNetworkResponse.Error.Message}
	}
	return &appmessage.GetCurrentNetworkResponseMessage{
		CurrentNetwork: x.GetCurrentNetworkResponse.CurrentNetwork,
		Error:          err,
	}, nil
}

func (x *KaspadMessage_GetCurrentNetworkResponse) fromAppMessage(message *appmessage.GetCurrentNetworkResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetCurrentNetworkResponse = &GetCurrentNetworkResponseMessage{
		CurrentNetwork: message.CurrentNetwork,
		Error:          err,
	}
	return nil
}
