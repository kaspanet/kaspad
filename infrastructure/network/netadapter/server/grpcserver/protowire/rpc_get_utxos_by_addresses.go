package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetUtxosByAddressesRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetUTXOsByAddressesRequestMessage{
		Addresses: x.GetUtxosByAddressesRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_GetUtxosByAddressesRequest) fromAppMessage(message *appmessage.GetUTXOsByAddressesRequestMessage) error {
	x.GetUtxosByAddressesRequest = &GetUtxosByAddressesRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_GetUtxosByAddressesResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetUtxosByAddressesResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetUtxosByAddressesResponse.Error.Message}
	}
	entries := make([]*appmessage.UTXOsByAddressesEntry, len(x.GetUtxosByAddressesResponse.Entries))
	for i, entry := range x.GetUtxosByAddressesResponse.Entries {
		entries[i] = entry.toAppMessage()
	}
	return &appmessage.GetUTXOsByAddressesResponseMessage{
		Entries: entries,
		Error:   err,
	}, nil
}

func (x *KaspadMessage_GetUtxosByAddressesResponse) fromAppMessage(message *appmessage.GetUTXOsByAddressesResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	entries := make([]*UtxosByAddressesEntry, len(message.Entries))
	for i, entry := range message.Entries {
		entries[i] = &UtxosByAddressesEntry{}
		entries[i].fromAppMessage(entry)
	}
	x.GetUtxosByAddressesResponse = &GetUtxosByAddressesResponseMessage{
		Entries: entries,
		Error:   err,
	}
	return nil
}
