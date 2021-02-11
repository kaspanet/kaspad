package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_StopNotifyingUtxosChangedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.StopNotifyingUTXOsChangedRequestMessage{
		Addresses: x.StopNotifyingUtxosChangedRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_StopNotifyingUtxosChangedRequest) fromAppMessage(message *appmessage.StopNotifyingUTXOsChangedRequestMessage) error {
	x.StopNotifyingUtxosChangedRequest = &StopNotifyingUtxosChangedRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_StopNotifyingUtxosChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.StopNotifyingUtxosChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.StopNotifyingUtxosChangedResponse.Error.Message}
	}
	return &appmessage.StopNotifyingUTXOsChangedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_StopNotifyingUtxosChangedResponse) fromAppMessage(message *appmessage.StopNotifyingUTXOsChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StopNotifyingUtxosChangedResponse = &StopNotifyingUtxosChangedResponseMessage{
		Error: err,
	}
	return nil
}
