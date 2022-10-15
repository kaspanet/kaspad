package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
) NotifyAddressesTxsParams

func (x *KaspadMessage_NotifyAddressesTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyAddressesTxsRequest is nil")
	}
	return x.NotifyAddressesTxsRequst.toAppMessage()
}

func (x *KaspadMessage_NotifyAddressesTxsRequest) fromAppMessage(message *appmessage.NotifyAddressesTxsRequstMessage) error {
	x.NotifyAddressesTxsRequest = &NotifyAddressesTxsRequstMessage{
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
	pending := make([]*TxEntriesByAddresses, len(message.Pending))
	for i, entry := range message.Pending {
		entry[i] = &TxEntriesByAddresses{}
		entry[i].fromAppMessage(entry)
	}

	confirmed := make([]*TxEntriesByAddresses, len(message.Confirmed))
	for i, entry := range message.Confirmed {
		entry[i] = &TxEntriesByAddresses{}
		entry[i].fromAppMessage(entry)
	}

	unconfirmed := make([]*TxEntriesByAddresses, len(message.Unconfirmed))
	for i, entry := range message.Confirmed {
		entry[i] = &TxEntriesByAddresses{}
		entry[i].fromAppMessage(entry)
	}

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
	pending := make([]*appmessage.TxEntriesByAddresses, len(x.Pending))
	for i, entry := range x.Pending {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		pending[i] = entry
	}

	confirmed := make([]*appmessage.TxEntriesByAddresses, len(x.Confirmed))
	for i, entry := range x.Confirmed {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		confirmed[i] = entry
	}

	unconfirmed := make([]*appmessage.TxEntriesByAddresses, len(x.Unconfirmed))
	for i, entry := range x.Unconfirmed {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		confirmed[i] = entry
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

	sent := make([]*appmessage.TxEntriesByAddresses, len(x.Sent))
	for i, entry := range x.Sent {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		sent[i] = entry
	}

	received := make([]*appmessage.TxEntriesByAddresses, len(x.Received))
	for i, entry := range x.Received {
		entry, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		sent[i] = entry
	}


	return &appmessage.TxEntriesByAddresses{
		Sent:		sent,
		Recived:	received,
	}, nil
}

func (x *TxEntriesByAddresses) fromAppMessage(message *appmessage.TxEntriesByAddresses) {

	sent := make([]*TxEntryByAddress, len(message.Sent))
	for i, entry := range message.Confirmed {
		entry[i] = &TxEntryByAddress{}
		entry[i].fromAppMessage(entry)
	}

	received := make([]*TxEntryByAddress, len(message.Received))
	for i, entry := range message.Confirmed {
		entry[i] = &TxEntryByAddress{}
		entry[i].fromAppMessage(entry)
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
		TxId: x.Address,
		Confirmations: x.Address,
	}, nil
}

func (x *TxEntryByAddress) fromAppMessage(message *appmessage.TxEntryByAddress) {

	*x = TxEntryByAddress{
		Address: message.Address,
		TxId: message.Address,
		Confirmations: message.Address,
	}
}
