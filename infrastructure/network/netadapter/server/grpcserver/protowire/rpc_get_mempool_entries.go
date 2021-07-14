package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetMempoolEntriesRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntriesRequest is nil")
	}
	return &appmessage.GetMempoolEntriesRequestMessage{}, nil
}

func (x *KaspadMessage_GetMempoolEntriesRequest) fromAppMessage(_ *appmessage.GetMempoolEntriesRequestMessage) error {
	x.GetMempoolEntriesRequest = &GetMempoolEntriesRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetMempoolEntriesResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntriesResponse is nil")
	}
	return x.GetMempoolEntriesResponse.toAppMessage()
}

func (x *KaspadMessage_GetMempoolEntriesResponse) fromAppMessage(message *appmessage.GetMempoolEntriesResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	entries, err := mempoolEntriesAppMessagesToProtos(message.Entries)
	if err != nil {
		return err
	}
	x.GetMempoolEntriesResponse = &GetMempoolEntriesResponseMessage{
		Entries: entries,
		Error:   rpcErr,
	}
	return nil
}

func (x *GetMempoolEntriesResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetMempoolEntriesResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.Entries) != 0 {
		return nil, errors.New("GetMempoolEntriesResponseMessage contains both an error and a response")
	}

	entries, err := mempoolEntriesProtosToAppMessages(x.Entries)
	if err != nil {
		return nil, err
	}

	return &appmessage.GetMempoolEntriesResponseMessage{
		Entries: entries,
		Error:   rpcErr,
	}, nil
}
