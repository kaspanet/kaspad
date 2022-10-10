package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxsRequest")
	}
	return x.GetTxsRequest.toAppMessage()
}

func (x *KaspadMessage_GetTxsRequest) fromAppMessage(message *appmessage.GetTxsRequestMessage) error {
	x.GetTxsRequest = &GetTxsRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *GetTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxsRequestMessage is nil")
	}
	return &appmessage.GetTxsRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_GetTxsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxsResponse is nil")
	}
	return x.GetTxsResponse.toAppMessage()
}

func (x *KaspadMessage_GetTxsResponse) fromAppMessage(message *appmessage.GetTxsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}

	rpcTransactions := make([]*RpcTransaction, len(message.Transactions))
	for i := range rpcTransactions {
		rpcTransactions[i].fromAppMessage(message.Transactions[i])
	}

	x.GetTxsResponse = &GetTxsResponseMessage{
		Transactions: rpcTransactions,

		Error: rpcErr,
	}
	return nil
}

func (x *GetTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Transactions != nil {
		return nil, errors.New("GetTxsResponseMessage contains both an error and a response")
	}

	appTransactions := make([]*appmessage.RPCTransaction, len(x.Transactions))
	for i := range appTransactions {
		appTransactions[i], err = x.Transactions[i].toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetTxsResponseMessage{
		Transactions: appTransactions,
		Error:        rpcErr,
	}, nil
}
