package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetMempoolEntriesByAddressesRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntriesRequest is nil")
	}
	return x.toAppMessage()
}

func (x *KaspadMessage_GetMempoolEntriesByAddressesRequest) fromAppMessage(message *appmessage.GetMempoolEntriesByAddressesRequestMessage) error {
	x.GetMempoolEntriesByAddressesRequest = &GetMempoolEntriesByAddressesRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_GetMempoolEntriesByAddressesResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntriesByAddressesResponse is nil")
	}
	return x.GetMempoolEntriesByAddressesResponse.toAppMessage()
}

func (x *KaspadMessage_GetMempoolEntriesByAddressesResponse) fromAppMessage(message *appmessage.GetMempoolEntriesByAddressesResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	entries := make([]*MempoolEntryByAddress, len(message.Entries))
	for i, entry := range message.Entries {
		err := entries[i].fromAppMessage(entry)
		if err != nil {
			return err
		}
	}
	x.GetMempoolEntriesByAddressesResponse = &GetMempoolEntriesByAddressesResponseMessage{
		Entries: entries,
		Error:   rpcErr,
	}
	return nil
}

func (x *GetMempoolEntriesByAddressesResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetMempoolEntriesResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.Entries) != 0 {
		return nil, errors.New("GetMempoolEntriesByAddressesResponseMessage contains both an error and a response")
	}
	entries := make([]*appmessage.MempoolEntryByAddress, len(x.Entries))
	for i, entry := range x.Entries {
		entries[i], err = entry.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetMempoolEntriesByAddressesResponseMessage{
		Entries: entries,
		Error:   rpcErr,
	}, nil
}

func (x *MempoolEntryByAddress) toAppMessage() (*appmessage.MempoolEntryByAddress, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "MempoolEntry is nil")
	}

	var err error

	sending := make([]*appmessage.MempoolEntry, len(x.Sending))
	for i, mempoolEntry := range x.Sending{
		sending[i], err = mempoolEntry.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	receiving := make([]*appmessage.MempoolEntry, len(x.Receiving))
	for i, mempoolEntry := range x.Receiving{
		receiving[i], err = mempoolEntry.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.MempoolEntryByAddress{
		Address:     x.Address,
		Sending:	sending,
		Receiving:	receiving,
	}, nil
}

func (x *MempoolEntryByAddress) fromAppMessage(message *appmessage.MempoolEntryByAddress) error {
	
	sending := make([]*MempoolEntry, len(message.Sending))
	for i, mempoolEntry := range message.Sending{
		sending[i].fromAppMessage(mempoolEntry)
	} 
	receiving := make([]*MempoolEntry, len(message.Receiving))
	for i, mempoolEntry := range message.Receiving{
		receiving[i].fromAppMessage(mempoolEntry)
	}

	*x = MempoolEntryByAddress{
		Address:	message.Address,
		Sending:	sending,
		Receiving:	receiving,
	}

	return nil
}
