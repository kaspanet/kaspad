package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetAcceptingBlocksOfTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlocksOfTxsRequest")
	}
	return x.GetAcceptingBlocksOfTxsRequest.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlocksOfTxsRequest) fromAppMessage(message *appmessage.GetAcceptingBlocksOfTxsRequestMessage) error {
	x.GetAcceptingBlocksOfTxsRequest = &GetAcceptingBlocksOfTxsRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *GetAcceptingBlocksOfTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlocksOfTxsRequestMessage is nil")
	}
	return &appmessage.GetAcceptingBlocksOfTxsRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_GetAcceptingBlocksOfTxsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlocksOfTxsResponse is nil")
	}
	return x.GetAcceptingBlocksOfTxsResponse.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlocksOfTxsResponse) fromAppMessage(message *appmessage.GetAcceptingBlocksOfTxsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}

	rpcTxIDBlockPairs := make([]*TxIDBlockPair, len(message.TxIDBlockPairs))
	for i := range rpcTxIDBlockPairs {
		err := rpcTxIDBlockPairs[i].fromAppMessage(message.TxIDBlockPairs[i])
		if err != nil {
			return err
		}
	}
	
	x.GetAcceptingBlocksOfTxsResponse = &GetAcceptingBlocksOfTxsResponseMessage{
		TxIDBlockPairs: rpcTxIDBlockPairs,

		Error: rpcErr,
	}
	return nil
}

func (x *GetAcceptingBlocksOfTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlocksOfTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.TxIDBlockPairs != nil {
		return nil, errors.New("GetAcceptingBlocksOfTxsResponseMessage contains both an error and a response")
	}

	appTxIDBlockPairs := make([]*appmessage.TxIDBlockPair, len(x.TxIDBlockPairs))
	for i := range appTxIDBlockPairs {
		appTxIDBlockPairs[i], err = x.TxIDBlockPairs[i].toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetAcceptingBlocksOfTxsResponseMessage{
		TxIDBlockPairs:  appTxIDBlockPairs,
		Error: rpcErr,
	}, nil
}

func (x *TxIDBlockPair) toAppMessage() (*appmessage.TxIDBlockPair, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TxIDBlockPair is nil")
	}

	appBlock, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.TxIDBlockPair{
		TxID:  x.TxID,
		Block: *appBlock,
	}, nil
}

func (x *TxIDBlockPair) fromAppMessage(message *appmessage.TxIDBlockPair) error {
	
	rpcBlock := &RpcBlock{}
	
	err := rpcBlock.fromAppMessage(&message.Block)
	if err != nil {
		return err
	}
	*x = TxIDBlockPair{
		TxID: message.TxID,
		Block: rpcBlock,

	}
	return nil
}
