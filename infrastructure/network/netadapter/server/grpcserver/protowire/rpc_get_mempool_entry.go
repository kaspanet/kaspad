package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetMempoolEntryRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetMempoolEntryRequestMessage{
		TxID: x.GetMempoolEntryRequest.TxId,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntryRequest) fromAppMessage(message *appmessage.GetMempoolEntryRequestMessage) error {
	x.GetMempoolEntryRequest = &GetMempoolEntryRequestMessage{
		TxId: message.TxID,
	}
	return nil
}

func (x *KaspadMessage_GetMempoolEntryResponse) toAppMessage() (appmessage.Message, error) {
	var rpcErr *appmessage.RPCError
	if x.GetMempoolEntryResponse.Error != nil {
		rpcErr = &appmessage.RPCError{Message: x.GetMempoolEntryResponse.Error.Message}
	}
	var entry *appmessage.MempoolEntry
	if x.GetMempoolEntryResponse.Entry != nil {
		var err error
		entry, err = x.GetMempoolEntryResponse.Entry.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.GetMempoolEntryResponseMessage{
		Entry: entry,
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntryResponse) fromAppMessage(message *appmessage.GetMempoolEntryResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	entry := new(MempoolEntry)
	if message.Entry != nil {
		err := entry.fromAppMessage(message.Entry)
		if err != nil {
			return err
		}
	}
	x.GetMempoolEntryResponse = &GetMempoolEntryResponseMessage{
		Entry: entry,
		Error: rpcErr,
	}
	return nil
}

func (x *MempoolEntry) toAppMessage() (*appmessage.MempoolEntry, error) {
	var txVerboseData *appmessage.TransactionVerboseData
	if x.TransactionVerboseData != nil {
		var err error
		txVerboseData, err = x.TransactionVerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.MempoolEntry{
		Fee:                    x.Fee,
		TransactionVerboseData: txVerboseData,
	}, nil
}

func (x *MempoolEntry) fromAppMessage(message *appmessage.MempoolEntry) error {
	var txVerboseData *TransactionVerboseData
	if message.TransactionVerboseData != nil {
		txVerboseData = new(TransactionVerboseData)
		err := txVerboseData.fromAppMessage(message.TransactionVerboseData)
		if err != nil {
			return err
		}
	}
	*x = MempoolEntry{
		Fee:                    message.Fee,
		TransactionVerboseData: txVerboseData,
	}
	return nil
}
