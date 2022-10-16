package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyAddressesTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyAddressesTxsRequest is nil")
	}
	return x.NotifyAddressesTxsRequest.toAppMessage()
}

func (x *KaspadMessage_NotifyAddressesTxsRequest) fromAppMessage(message *appmessage.NotifyAddressesTxsRequestMessage) error {
	x.NotifyAddressesTxsRequest = &NotifyAddressesTxsRequestMessage{
		Addresses : message.Addresses,
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
		IncludeSending: message.IncludeSending,
		IncludeReceiving: message.IncludeReceiving,
	}
	return nil
}

func (x *NotifyAddressesTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyAddressesTxsRequestMessage is nil")
	}
	return &appmessage.NotifyAddressesTxsRequestMessage{
		Addresses : x.Addresses,
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
		IncludeSending: x.IncludeSending,
		IncludeReceiving: x.IncludeReceiving,
	}, nil
}

func (x *KaspadMessage_NotifyAddressesTxsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyAddressesTxsResponseMessage is nil")
	}
	return x.NotifyAddressesTxsResponse.toAppMessage()
}

func (x *KaspadMessage_NotifyAddressesTxsResponse) fromAppMessage(message *appmessage.NotifyAddressesTxsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyAddressesTxsResponse = &NotifyAddressesTxsResponseMessage{
		Error: err,
	}
	return nil
}

func (x *NotifyAddressesTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyAddressesTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyAddressesTxsResponseMessage{
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_AddressesTxsNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_AddressesTxsNotification is nil")
	}
	return x.AddressesTxsNotification.toAppMessage()
}

func (x *KaspadMessage_AddressesTxsNotification) fromAppMessage(message *appmessage.AddressesTxsNotificationMessage) error {

	pending := &TxEntriesByAddresses{}
	pending.fromAppMessage(message.Pending)

	confirmed := &TxEntriesByAddresses{}
	confirmed.fromAppMessage(message.Confirmed)

	unconfirmed := &TxEntriesByAddresses{}
	unconfirmed.fromAppMessage(message.Unconfirmed)

	x.AddressesTxsNotification = &AddressesTxsNotificationMessage{
		RequiredConfirmations: message.RequiredConfirmations,
		Pending: pending,
		Confirmed: confirmed,
		Unconfirmed: unconfirmed,
	}
	return nil
}

func (x *AddressesTxsNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "AddressesTxsNotificationMessage is nil")
	}
	pending, err := x.Pending.toAppMessage()
	if err != nil {
		return nil, err
	}

	confirmed, err := x.Pending.toAppMessage()
	if err != nil {
		return nil, err
	}

	unconfirmed, err := x.Unconfirmed.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.AddressesTxsNotificationMessage{
		RequiredConfirmations: x.RequiredConfirmations,
		Pending: pending,
		Confirmed: confirmed,
		Unconfirmed: unconfirmed,
	}, nil
}


func (x *TxEntriesByAddresses) toAppMessage() (*appmessage.TxEntriesByAddresses, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TxEntriesByAddresses is nil")
	}

	sent := make([]*appmessage.TxEntryByAddress, len(x.Sent))
	for i, entry := range x.Sent {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		sent[i] = entry
	}

	received := make([]*appmessage.TxEntryByAddress, len(x.Received))
	for i, entry := range x.Received {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		received[i] = entry
	}


	return &appmessage.TxEntriesByAddresses{
		Sent:		sent,
		Received:		received,
	}, nil
}

func (x *TxEntriesByAddresses) fromAppMessage(message *appmessage.TxEntriesByAddresses) {

	sent := make([]*TxEntryByAddress, len(message.Sent))
	for i, entry := range message.Sent {
		sent[i] = &TxEntryByAddress{}
		sent[i].fromAppMessage(entry)
	}

	received := make([]*TxEntryByAddress, len(message.Received))
	for i, entry := range message.Received {
		received[i] = &TxEntryByAddress{}
		received[i].fromAppMessage(entry)
	}

	*x = TxEntriesByAddresses{
		Sent:		sent,
		Received:	received,
	}
}

func (x *TxEntryByAddress) toAppMessage() (*appmessage.TxEntryByAddress, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TxEntryByAddress is nil")
	}

	return &appmessage.TxEntryByAddress{
		Address: x.Address,
		TxID: x.TxId,
		Confirmations: x.Confirmations,
	}, nil
}

func (x *TxEntryByAddress) fromAppMessage(message *appmessage.TxEntryByAddress) {

	*x = TxEntryByAddress{
		Address: message.Address,
		TxId: message.TxID,
		Confirmations: message.Confirmations,
	}
}
