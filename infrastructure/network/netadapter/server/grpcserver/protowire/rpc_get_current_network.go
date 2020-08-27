package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetCurrentNetworkRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetCurrentNetworkRequestMessage{}, nil
}

func (x *KaspadMessage_GetCurrentNetworkRequest) fromAppMessage(_ *appmessage.GetCurrentNetworkRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetCurrentNetworkResponse) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetCurrentNetworkResponseMessage{
		CurrentNetwork: x.GetCurrentNetworkResponse.CurrentNetwork,
	}, nil
}

func (x *KaspadMessage_GetCurrentNetworkResponse) fromAppMessage(message *appmessage.GetCurrentNetworkResponseMessage) error {
	x.GetCurrentNetworkResponse = &GetCurrentNetworkResponseMessage{
		CurrentNetwork: message.CurrentNetwork,
	}
	return nil
}
