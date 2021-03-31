package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetMempoolEntryRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntryRequest is nil")
	}
	return x.GetMempoolEntryRequest.toAppMessage()
}

func (x *KaspadMessage_GetMempoolEntryRequest) fromAppMessage(message *appmessage.GetMempoolEntryRequestMessage) error {
	x.GetMempoolEntryRequest = &GetMempoolEntryRequestMessage{
		TxId: message.TxID,
	}
	return nil
}

func (x *GetMempoolEntryRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetMempoolEntryRequestMessage is nil")
	}
	return &appmessage.GetMempoolEntryRequestMessage{
		TxID: x.TxId,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntryResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetMempoolEntryResponse is nil")
	}
	return x.GetMempoolEntryResponse.toAppMessage()
}

func (x *KaspadMessage_GetMempoolEntryResponse) fromAppMessage(message *appmessage.GetMempoolEntryResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	var entry *MempoolEntry
	if message.Entry != nil {
		entry = new(MempoolEntry)
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

func (x *GetMempoolEntryResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetMempoolEntryResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	entry, err := x.Entry.toAppMessage()
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && entry != nil {
		return nil, errors.New("GetMempoolEntryResponseMessage contains both an error and a response")
	}

	return &appmessage.GetMempoolEntryResponseMessage{
		Entry: entry,
		Error: rpcErr,
	}, nil
}

func (x *MempoolEntry) toAppMessage() (*appmessage.MempoolEntry, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "MempoolEntry is nil")
	}
	transaction, err := x.Transaction.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.MempoolEntry{
		Fee:         x.Fee,
		Transaction: transaction,
	}, nil
}

func (x *MempoolEntry) fromAppMessage(message *appmessage.MempoolEntry) error {
	var transaction *RpcTransaction
	if message.Transaction != nil {
		transaction = new(RpcTransaction)
		transaction.fromAppMessage(message.Transaction)
	}
	*x = MempoolEntry{
		Fee:         message.Fee,
		Transaction: transaction,
	}
	return nil
}
