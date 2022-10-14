package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetIncludingBlockHashesOfTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlockHashesOfTxsRequest")
	}
	return x.GetIncludingBlockHashesOfTxsRequest.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlockHashesOfTxsRequest) fromAppMessage(message *appmessage.GetIncludingBlockHashesOfTxsRequestMessage) error {
	x.GetIncludingBlockHashesOfTxsRequest = &GetIncludingBlockHashesOfTxsRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *GetIncludingBlockHashesOfTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlockHashesOfTxsRequestMessage is nil")
	}
	return &appmessage.GetIncludingBlockHashesOfTxsRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_GetIncludingBlockHashesOfTxsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludinglockHashesOfTxsResponse is nil")
	}
	return x.GetIncludingBlockHashesOfTxsResponse.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlockHashesOfTxsResponse) fromAppMessage(message *appmessage.GetIncludingBlockHashesOfTxsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}

	rpcTxIDBlockHashPairs := make([]*RpcTxIDBlockHashPair, len(message.TxIDBlockHashPairs))
	for i := range rpcTxIDBlockHashPairs {
		rpcTxIDBlockHashPairs[i] = &RpcTxIDBlockHashPair{}
		rpcTxIDBlockHashPairs[i].fromAppMessage(message.TxIDBlockHashPairs[i])
	}

	x.GetIncludingBlockHashesOfTxsResponse = &GetIncludingBlockHashesOfTxsResponseMessage{
		TxIDBlockHashPairs: rpcTxIDBlockHashPairs,

		Error: rpcErr,
	}
	return nil
}

func (x *GetIncludingBlockHashesOfTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlockHashesOfTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.TxIDBlockHashPairs != nil {
		return nil, errors.New("GetIncludingBlockHashesfTxsResponseMessage contains both an error and a response")
	}

	appTxIDBlockHashPairs := make([]*appmessage.TxIDBlockHashPair, len(x.TxIDBlockHashPairs))
	for i := range appTxIDBlockHashPairs {
		appTxIDBlockHashPairs[i], err = x.TxIDBlockHashPairs[i].toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetIncludingBlockHashesOfTxsResponseMessage{
		TxIDBlockHashPairs: appTxIDBlockHashPairs,
		Error:              rpcErr,
	}, nil
}
