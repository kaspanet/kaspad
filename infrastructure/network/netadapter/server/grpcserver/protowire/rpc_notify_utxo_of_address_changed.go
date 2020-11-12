package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyUTXOOfAddressChangedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyUTXOOfAddressChangedRequestMessage{
		Addresses: x.NotifyUTXOOfAddressChangedRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_NotifyUTXOOfAddressChangedRequest) fromAppMessage(message *appmessage.NotifyUTXOOfAddressChangedRequestMessage) error {
	x.NotifyUTXOOfAddressChangedRequest = &NotifyUTXOOfAddressChangedRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_NotifyUTXOOfAddressChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyUTXOOfAddressChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyUTXOOfAddressChangedResponse.Error.Message}
	}
	return &appmessage.NotifyUTXOOfAddressChangedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyUTXOOfAddressChangedResponse) fromAppMessage(message *appmessage.NotifyUTXOOfAddressChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyUTXOOfAddressChangedResponse = &NotifyUTXOOfAddressChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_UtxoOfAddressChangedNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.UTXOOfAddressChangedNotificationMessage{
		ChangedAddresses: x.UtxoOfAddressChangedNotification.ChangedAddresses,
	}, nil
}

func (x *KaspadMessage_UtxoOfAddressChangedNotification) fromAppMessage(message *appmessage.UTXOOfAddressChangedNotificationMessage) error {
	x.UtxoOfAddressChangedNotification = &UTXOOfAddressChangedNotificationMessage{
		ChangedAddresses: message.ChangedAddresses,
	}
	return nil
}
