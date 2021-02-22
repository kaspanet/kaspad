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

func (x *KaspadMessage_StopNotifyPruningPointUTXOSetOverrideRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.StopNotifyPruningPointUTXOSetOverrideRequestMessage{}, nil
}

func (x *KaspadMessage_StopNotifyPruningPointUTXOSetOverrideRequest) fromAppMessage(_ *appmessage.StopNotifyPruningPointUTXOSetOverrideRequestMessage) error {
	x.StopNotifyPruningPointUTXOSetOverrideRequest = &StopNotifyPruningPointUTXOSetOverrideRequestMessage{}
	return nil
}

func (x *KaspadMessage_StopNotifyPruningPointUTXOSetOverrideResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.StopNotifyPruningPointUTXOSetOverrideResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.StopNotifyPruningPointUTXOSetOverrideResponse.Error.Message}
	}
	return &appmessage.StopNotifyPruningPointUTXOSetOverrideResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_StopNotifyPruningPointUTXOSetOverrideResponse) fromAppMessage(
	message *appmessage.StopNotifyPruningPointUTXOSetOverrideResponseMessage) error {

	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StopNotifyPruningPointUTXOSetOverrideResponse = &StopNotifyPruningPointUTXOSetOverrideResponseMessage{
		Error: err,
	}
	return nil
}
