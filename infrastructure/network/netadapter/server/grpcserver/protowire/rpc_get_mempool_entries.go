package protowire

import (
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_GetMempoolEntriesRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntriesRequest is nil")
	}
	return x.GetMempoolEntriesRequest.toAppMessage()
}

func (x *KaspadMessage_GetMempoolEntriesRequest) fromAppMessage(message *appmessage.GetMempoolEntriesRequestMessage) error {
	x.GetMempoolEntriesRequest = &GetMempoolEntriesRequestMessage{
		IncludeOrphanPool:     message.IncludeOrphanPool,
		FilterTransactionPool: message.FilterTransactionPool,
	}
	return nil
}

func (x *GetMempoolEntriesRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetMempoolEntryRequestMessage is nil")
	}
	return &appmessage.GetMempoolEntriesRequestMessage{
		IncludeOrphanPool:     x.IncludeOrphanPool,
		FilterTransactionPool: x.FilterTransactionPool,
	}, nil
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
	entries := make([]*MempoolEntry, len(message.Entries))
	for i, entry := range message.Entries {
		entries[i] = new(MempoolEntry)
		err := entries[i].fromAppMessage(entry)
		if err != nil {
			return err
		}
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
	entries := make([]*appmessage.MempoolEntry, len(x.Entries))
	for i, entry := range x.Entries {
		entries[i], err = entry.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetMempoolEntriesResponseMessage{
		Entries: entries,
		Error:   rpcErr,
	}, nil
}
