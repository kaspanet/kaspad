package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetTxRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxRequest")
	}
	return x.GetTxRequest.toAppMessage()
}

func (x *KaspadMessage_GetTxRequest) fromAppMessage(message *appmessage.GetTxRequestMessage) error {
	x.GetTxRequest = &GetTxRequestMessage{
		TxID: message.TxID,
	}
	return nil
}

func (x *GetTxRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxRequestMessage is nil")
	}
	return &appmessage.GetTxRequestMessage{
		TxID: x.TxID,
	}, nil
}

func (x *KaspadMessage_GetTxResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxResponse is nil")
	}
	return x.GetTxResponse.toAppMessage()
}

func (x *KaspadMessage_GetTxResponse) fromAppMessage(message *appmessage.GetTxResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	rpcTransaction := &RpcTransaction{}
	rpcTransaction.fromAppMessage(message.Transaction)

	x.GetTxResponse = &GetTxResponseMessage{
		Transaction: rpcTransaction,

		Error: rpcErr,
	}
	return nil
}

func (x *GetTxResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Transaction != nil {
		return nil, errors.New("GetTxResponseMessage contains both an error and a response")
	}

	appTransaction, err := x.Transaction.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.GetTxResponseMessage{
		Transaction: appTransaction,
		Error:       rpcErr,
	}, nil
}
