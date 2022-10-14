package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetIncludingBlocksOfTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlocksOfTxsRequest")
	}
	return x.GetIncludingBlocksOfTxsRequest.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlocksOfTxsRequest) fromAppMessage(message *appmessage.GetIncludingBlocksOfTxsRequestMessage) error {
	x.GetIncludingBlocksOfTxsRequest = &GetIncludingBlocksOfTxsRequestMessage{
		TxIDs:               message.TxIDs,
		IncludeTransactions: message.IncludeTransactions,
	}
	return nil
}

func (x *GetIncludingBlocksOfTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlocksOfTxsRequestMessage is nil")
	}
	return &appmessage.GetIncludingBlocksOfTxsRequestMessage{
		TxIDs:               x.TxIDs,
		IncludeTransactions: x.IncludeTransactions,
	}, nil
}

func (x *KaspadMessage_GetIncludingBlocksOfTxsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlocksOfTxsResponse is nil")
	}
	return x.GetIncludingBlocksOfTxsResponse.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlocksOfTxsResponse) fromAppMessage(message *appmessage.GetIncludingBlocksOfTxsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}

	rpcTxIDBlockPairs := make([]*RpcTxIDBlockPair, len(message.TxIDBlockPairs))
	for i := range rpcTxIDBlockPairs {
		rpcTxIDBlockPairs[i] = &RpcTxIDBlockPair{}
		err := rpcTxIDBlockPairs[i].fromAppMessage(message.TxIDBlockPairs[i])
		if err != nil {
			return err
		}
	}

	x.GetIncludingBlocksOfTxsResponse = &GetIncludingBlocksOfTxsResponseMessage{
		TxIDBlockPairs: rpcTxIDBlockPairs,

		Error: rpcErr,
	}
	return nil
}

func (x *GetIncludingBlocksOfTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlocksOfTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.TxIDBlockPairs != nil {
		return nil, errors.New("GetIncludingBlocksOfTxsResponseMessage contains both an error and a response")
	}

	appTxIDBlockPairs := make([]*appmessage.TxIDBlockPair, len(x.TxIDBlockPairs))
	for i := range appTxIDBlockPairs {
		appTxIDBlockPairs[i], err = x.TxIDBlockPairs[i].toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetIncludingBlocksOfTxsResponseMessage{
		TxIDBlockPairs: appTxIDBlockPairs,
		Error:          rpcErr,
	}, nil
}
