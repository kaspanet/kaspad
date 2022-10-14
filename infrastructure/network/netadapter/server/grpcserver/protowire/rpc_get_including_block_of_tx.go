package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetIncludingBlockOfTxRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlockOfTxRequest")
	}
	return x.GetIncludingBlockOfTxRequest.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlockOfTxRequest) fromAppMessage(message *appmessage.GetIncludingBlockOfTxRequestMessage) error {
	x.GetIncludingBlockOfTxRequest = &GetIncludingBlockOfTxRequestMessage{
		TxID:                message.TxID,
		IncludeTransactions: message.IncludeTransactions,
	}
	return nil
}

func (x *GetIncludingBlockOfTxRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlockOfTxRequestMessage is nil")
	}
	return &appmessage.GetIncludingBlockOfTxRequestMessage{
		TxID:                x.TxID,
		IncludeTransactions: x.IncludeTransactions,
	}, nil
}

func (x *KaspadMessage_GetIncludingBlockOfTxResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlockOfTxResponse is nil")
	}
	return x.GetIncludingBlockOfTxResponse.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlockOfTxResponse) fromAppMessage(message *appmessage.GetIncludingBlockOfTxResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	rpcBlock := &RpcBlock{}

	err := rpcBlock.fromAppMessage(message.Block)
	if err != nil {
		return err
	}
	x.GetIncludingBlockOfTxResponse = &GetIncludingBlockOfTxResponseMessage{
		Block: rpcBlock,

		Error: rpcErr,
	}
	return nil
}

func (x *GetIncludingBlockOfTxResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlockOfTxResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Block != nil {
		return nil, errors.New("GetIncludingBlockOfTxResponseMessage contains both an error and a response")
	}

	appBlock, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.GetIncludingBlockOfTxResponseMessage{
		Block: appBlock,
		Error: rpcErr,
	}, nil
}
