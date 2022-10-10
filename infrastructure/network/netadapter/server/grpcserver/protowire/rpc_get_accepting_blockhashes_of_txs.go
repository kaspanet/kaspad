package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetAcceptingBlockHashesOfTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlockHashesOfTxsRequest")
	}
	return x.GetAcceptingBlockHashesOfTxsRequest.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlockHashesOfTxsRequest) fromAppMessage(message *appmessage.GetAcceptingBlockHashesOfTxsRequestMessage) error {
	x.GetAcceptingBlockHashesOfTxsRequest = &GetAcceptingBlockHashesOfTxsRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *GetAcceptingBlockHashesOfTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlockHashesOfTxsRequestMessage is nil")
	}
	return &appmessage.GetAcceptingBlockHashesOfTxsRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_GetAcceptingBlockHashesOfTxsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlockHashesOfTxsResponse is nil")
	}
	return x.GetAcceptingBlockHashesOfTxsResponse.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlockHashesOfTxsResponse) fromAppMessage(message *appmessage.GetAcceptingBlockHashesOfTxsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}

	rpcTxIDBlockHashPairs := make([]*TxIDBlockHashPair, len(message.TxIDBlockHashPairs))
	for i := range rpcTxIDBlockHashPairs {
		rpcTxIDBlockHashPairs[i].fromAppMessage(message.TxIDBlockHashPairs[i])
	}

	x.GetAcceptingBlockHashesOfTxsResponse = &GetAcceptingBlockHashesOfTxsResponseMessage{
		TxIDBlockHashPairs: rpcTxIDBlockHashPairs,

		Error: rpcErr,
	}
	return nil
}

func (x *GetAcceptingBlockHashesOfTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlockHashesOfTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.TxIDBlockHashPairs != nil {
		return nil, errors.New("GetAcceptingBlockHashesfTxsResponseMessage contains both an error and a response")
	}

	appTxIDBlockHashPairs := make([]*appmessage.TxIDBlockHashPair, len(x.TxIDBlockHashPairs))
	for i := range appTxIDBlockHashPairs {
		appTxIDBlockHashPairs[i], err = x.TxIDBlockHashPairs[i].toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetAcceptingBlockHashesOfTxsResponseMessage{
		TxIDBlockHashPairs: appTxIDBlockHashPairs,
		Error:              rpcErr,
	}, nil
}

func (x *TxIDBlockHashPair) toAppMessage() (*appmessage.TxIDBlockHashPair, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TxIDBlockHashPair is nil")
	}

	return &appmessage.TxIDBlockHashPair{
		TxID: x.TxID,
		Hash: x.Hash,
	}, nil
}

func (x *TxIDBlockHashPair) fromAppMessage(message *appmessage.TxIDBlockHashPair) {

	*x = TxIDBlockHashPair{
		TxID: message.TxID,
		Hash: message.Hash,
	}
}
