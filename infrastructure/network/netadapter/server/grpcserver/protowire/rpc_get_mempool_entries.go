package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetMempoolEntriesRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetMempoolEntriesRequestMessage{}, nil
}

func (x *KaspadMessage_GetMempoolEntriesRequest) fromAppMessage(_ *appmessage.GetMempoolEntriesRequestMessage) error {
	x.GetMempoolEntriesRequest = &GetMempoolEntriesRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetMempoolEntriesResponse) toAppMessage() (appmessage.Message, error) {
	var rpcErr *appmessage.RPCError
	if x.GetMempoolEntriesResponse.Error != nil {
		rpcErr = &appmessage.RPCError{Message: x.GetMempoolEntriesResponse.Error.Message}
	}
	entries := make([]*appmessage.MempoolEntry, len(x.GetMempoolEntriesResponse.Entries))
	for i, entry := range x.GetMempoolEntriesResponse.Entries {
		var err error
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
