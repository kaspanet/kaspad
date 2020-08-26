package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetCurrentNetworkRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetCurrentNetworkRequestMessage{}, nil
}

func (x *KaspadMessage_GetCurrentNetworkRequest) fromAppMessage(_ *appmessage.GetCurrentNetworkRequestMessage) error {
	return nil
}
