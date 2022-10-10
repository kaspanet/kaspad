package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetAcceptingBlockOfTxRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlockOfTxRequest")
	}
	return x.GetAcceptingBlockOfTxRequest.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlockOfTxRequest) fromAppMessage(message *appmessage.GetAcceptingBlockOfTxRequestMessage) error {
	x.GetAcceptingBlockOfTxRequest = &GetAcceptingBlockOfTxRequestMessage{
		TxID:                message.TxID,
		IncludeTransactions: message.IncludeTransactions,
	}
	return nil
}

func (x *GetAcceptingBlockOfTxRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlockOfTxRequestMessage is nil")
	}
	return &appmessage.GetAcceptingBlockOfTxRequestMessage{
		TxID:                x.TxID,
		IncludeTransactions: x.IncludeTransactions,
	}, nil
}

func (x *KaspadMessage_GetAcceptingBlockOfTxResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlockOfTxResponse is nil")
	}
	return x.GetAcceptingBlockOfTxResponse.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlockOfTxResponse) fromAppMessage(message *appmessage.GetAcceptingBlockOfTxResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	rpcBlock := &RpcBlock{}

	err := rpcBlock.fromAppMessage(message.Block)
	if err != nil {
		return err
	}
	x.GetAcceptingBlockOfTxResponse = &GetAcceptingBlockOfTxResponseMessage{
		Block: rpcBlock,

		Error: rpcErr,
	}
	return nil
}

func (x *GetAcceptingBlockOfTxResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlockOfTxResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Block != nil {
		return nil, errors.New("GetAcceptingBlockOfTxResponseMessage contains both an error and a response")
	}

	appBlock, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.GetAcceptingBlockOfTxResponseMessage{
		Block: appBlock,
		Error: rpcErr,
	}, nil
}
