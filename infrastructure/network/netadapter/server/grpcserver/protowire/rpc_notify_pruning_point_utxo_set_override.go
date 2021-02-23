package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyPruningPointUTXOSetOverrideRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyPruningPointUTXOSetOverrideRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyPruningPointUTXOSetOverrideRequest) fromAppMessage(_ *appmessage.NotifyPruningPointUTXOSetOverrideRequestMessage) error {
	x.NotifyPruningPointUTXOSetOverrideRequest = &NotifyPruningPointUTXOSetOverrideRequestMessage{}
	return nil
}

func (x *KaspadMessage_NotifyPruningPointUTXOSetOverrideResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyPruningPointUTXOSetOverrideResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyPruningPointUTXOSetOverrideResponse.Error.Message}
	}
	return &appmessage.NotifyPruningPointUTXOSetOverrideResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyPruningPointUTXOSetOverrideResponse) fromAppMessage(message *appmessage.NotifyPruningPointUTXOSetOverrideResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyPruningPointUTXOSetOverrideResponse = &NotifyPruningPointUTXOSetOverrideResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_PruningPointUTXOSetOverrideNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.PruningPointUTXOSetOverrideNotificationMessage{}, nil
}

func (x *KaspadMessage_PruningPointUTXOSetOverrideNotification) fromAppMessage(_ *appmessage.PruningPointUTXOSetOverrideNotificationMessage) error {
	x.PruningPointUTXOSetOverrideNotification = &PruningPointUTXOSetOverrideNotificationMessage{}
	return nil
}

func (x *KaspadMessage_StopNotifyingPruningPointUTXOSetOverrideRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.StopNotifyingPruningPointUTXOSetOverrideRequestMessage{}, nil
}

func (x *KaspadMessage_StopNotifyingPruningPointUTXOSetOverrideRequest) fromAppMessage(_ *appmessage.StopNotifyingPruningPointUTXOSetOverrideRequestMessage) error {
	x.StopNotifyingPruningPointUTXOSetOverrideRequest = &StopNotifyingPruningPointUTXOSetOverrideRequestMessage{}
	return nil
}

func (x *KaspadMessage_StopNotifyingPruningPointUTXOSetOverrideResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.StopNotifyingPruningPointUTXOSetOverrideResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.StopNotifyingPruningPointUTXOSetOverrideResponse.Error.Message}
	}
	return &appmessage.StopNotifyingPruningPointUTXOSetOverrideResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_StopNotifyingPruningPointUTXOSetOverrideResponse) fromAppMessage(
	message *appmessage.StopNotifyingPruningPointUTXOSetOverrideResponseMessage) error {

	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StopNotifyingPruningPointUTXOSetOverrideResponse = &StopNotifyingPruningPointUTXOSetOverrideResponseMessage{
		Error: err,
	}
	return nil
}
