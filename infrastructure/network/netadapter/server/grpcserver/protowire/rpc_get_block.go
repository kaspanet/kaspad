package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBlockRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockRequest is nil")
	}
	return x.GetBlockRequest.toAppMessage()
}

func (x *GetBlockRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockRequestMessage is nil")
	}
	return &appmessage.GetBlockRequestMessage{
		Hash:                          x.Hash,
		IncludeTransactionVerboseData: x.IncludeTransactionVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlockRequest) fromAppMessage(message *appmessage.GetBlockRequestMessage) error {
	x.GetBlockRequest = &GetBlockRequestMessage{
		Hash:                          message.Hash,
		IncludeTransactionVerboseData: message.IncludeTransactionVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetBlockResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockResponse is nil")
	}
	return x.GetBlockResponse.toAppMessage()
}

func (x *GetBlockResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	var block *appmessage.RPCBlock
	// Return verbose data only if there's no error
	if rpcErr != nil && x.Block != nil {
		return nil, errors.New("GetBlockResponseMessage contains both an error and a response")
	}
	if rpcErr == nil {
		block, err = x.Block.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.GetBlockResponseMessage{
		Block: block,
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_GetBlockResponse) fromAppMessage(message *appmessage.GetBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	var block *RpcBlock
	if message.Block != nil {
		protoBlock := &RpcBlock{}
		err := protoBlock.fromAppMessage(message.Block)
		if err != nil {
			return err
		}
		block = protoBlock
	}
	x.GetBlockResponse = &GetBlockResponseMessage{
		Block: block,
		Error: err,
	}
	return nil
}
