package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetMempoolEntriesByAddressesRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntriesByAddressesRequest is nil")
	}
	return &appmessage.GetMempoolEntriesByAddressesRequestMessage{
		Addresses: x.GetMempoolEntriesByAddressesRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntriesByAddressesRequest) fromAppMessage(
	message *appmessage.GetMempoolEntriesByAddressesRequestMessage) error {

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

func (x *KaspadMessage_GetMempoolEntriesByAddressesResponse) fromAppMessage(
	message *appmessage.GetMempoolEntriesByAddressesResponseMessage) error {

	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	receivingEntries, err := mempoolEntriesAppMessagesToProtos(message.ReceivingEntries)
	if err != nil {
		return err
	}
	spendingEntries, err := mempoolEntriesAppMessagesToProtos(message.SpendingEntries)
	if err != nil {
		return err
	}
	x.GetMempoolEntriesByAddressesResponse = &GetMempoolEntriesByAddressesResponseMessage{
		ReceivingEntries: receivingEntries,
		SpendingEntries:  spendingEntries,
		Error:            rpcErr,
	}
	return nil
}

func (x *GetMempoolEntriesByAddressesResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetMempoolEntriesByAddressesResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && (len(x.SpendingEntries) != 0 || len(x.ReceivingEntries) != 0) {
		return nil, errors.New("GetMempoolEntriesByAddressesResponseMessage contains both an error and a response")
	}

	spendingEntries, err := mempoolEntriesProtosToAppMessages(x.SpendingEntries)
	if err != nil {
		return nil, err
	}
	receivingEntries, err := mempoolEntriesProtosToAppMessages(x.ReceivingEntries)
	if err != nil {
		return nil, err
	}

	return &appmessage.GetMempoolEntriesByAddressesResponseMessage{
		SpendingEntries:  spendingEntries,
		ReceivingEntries: receivingEntries,
		Error:            rpcErr,
	}, nil
}

func mempoolEntriesAppMessagesToProtos(appMessageEntries []*appmessage.MempoolEntry) ([]*MempoolEntry, error) {
	protos := make([]*MempoolEntry, len(appMessageEntries))
	for i, entry := range appMessageEntries {
		protos[i] = new(MempoolEntry)
		err := protos[i].fromAppMessage(entry)
		if err != nil {
			return nil, err
		}
	}
	return protos, nil
}

func mempoolEntriesProtosToAppMessages(protoEntries []*MempoolEntry) ([]*appmessage.MempoolEntry, error) {
	appMessages := make([]*appmessage.MempoolEntry, len(protoEntries))
	var err error
	for i, entry := range protoEntries {
		appMessages[i], err = entry.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return appMessages, nil
}
