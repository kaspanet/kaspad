package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBalancesByAddressesRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBalanceByAddressRequest is nil")
	}
	return x.GetBalancesByAddressesRequest.toAppMessage()
}

func (x *KaspadMessage_GetBalancesByAddressesRequest) fromAppMessage(message *appmessage.GetBalancesByAddressesRequestMessage) error {
	x.GetBalancesByAddressesRequest = &GetBalancesByAddressesRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *GetBalancesByAddressesRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressRequest is nil")
	}
	return &appmessage.GetBalancesByAddressesRequestMessage{
		Addresses: x.Addresses,
	}, nil
}

func (x *KaspadMessage_GetBalancesByAddressesResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressResponse is nil")
	}
	return x.GetBalancesByAddressesResponse.toAppMessage()
}

func (x *KaspadMessage_GetBalancesByAddressesResponse) fromAppMessage(message *appmessage.GetBalancesByAddressesResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	entries := make([]*BalancesByAddressEntry, len(message.Entries))
	for i, entry := range message.Entries {
		entries[i] = &BalancesByAddressEntry{}
		entries[i].fromAppMessage(entry)
	}
	x.GetBalancesByAddressesResponse = &GetBalancesByAddressesResponseMessage{
		Entries: entries,
		Error:   err,
	}
	return nil
}

func (x *GetBalancesByAddressesResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalancesByAddressesResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.Entries) != 0 {
		return nil, errors.New("GetBalancesByAddressesResponseMessage contains both an error and a response")
	}

	entries := make([]*appmessage.BalancesByAddressesEntry, len(x.Entries))
	for i, entry := range x.Entries {
		entryAsAppMessage, err := entry.toAppMessage()
		if err != nil {
			return nil, err
		}
		entries[i] = entryAsAppMessage
	}

	return &appmessage.GetBalancesByAddressesResponseMessage{
		Entries: entries,
		Error:   rpcErr,
	}, nil
}

func (x *BalancesByAddressEntry) toAppMessage() (*appmessage.BalancesByAddressesEntry, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BalancesByAddressesEntry is nil")
	}
	return &appmessage.BalancesByAddressesEntry{
		Address: x.Address,
		Balance: x.Balance,
	}, nil
}

func (x *BalancesByAddressEntry) fromAppMessage(message *appmessage.BalancesByAddressesEntry) {
	*x = BalancesByAddressEntry{
		Address: message.Address,
		Balance: message.Balance,
	}
}
