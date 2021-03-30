package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBlocksRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlocksRequest is nil")
	}
	return x.GetBlocksRequest.toAppMessage()
}

func (x *KaspadMessage_GetBlocksRequest) fromAppMessage(message *appmessage.GetBlocksRequestMessage) error {
	x.GetBlocksRequest = &GetBlocksRequestMessage{
		LowHash:                       message.LowHash,
		IncludeBlocks:                 message.IncludeBlocks,
		IncludeTransactionVerboseData: message.IncludeTransactionVerboseData,
	}
	return nil
}

func (x *GetBlocksRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlocksRequestMessage is nil")
	}
	return &appmessage.GetBlocksRequestMessage{
		LowHash:                       x.LowHash,
		IncludeBlocks:                 x.IncludeBlocks,
		IncludeTransactionVerboseData: x.IncludeTransactionVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlocksResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlocksResponse is nil")
	}
	return x.GetBlocksResponse.toAppMessage()
}

func (x *KaspadMessage_GetBlocksResponse) fromAppMessage(message *appmessage.GetBlocksResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlocksResponse = &GetBlocksResponseMessage{
		Error: err,
	}
	x.GetBlocksResponse.BlockHashes = message.BlockHashes
	x.GetBlocksResponse.Blocks = make([]*RpcBlock, len(message.Blocks))
	for i, block := range message.Blocks {
		protoBlock := &RpcBlock{}
		err := protoBlock.fromAppMessage(block)
		if err != nil {
			return err
		}
		x.GetBlocksResponse.Blocks[i] = protoBlock
	}
	return nil
}

func (x *GetBlocksResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlocksResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	// Return data only if there's no error
	if rpcErr != nil && len(x.Blocks) != 0 {
		return nil, errors.New("GetBlocksResponseMessage contains both an error and a response")
	}
	blocks := make([]*appmessage.RPCBlock, len(x.Blocks))
	for i, block := range x.Blocks {
		appMessageBlock, err := block.toAppMessage()
		if err != nil {
			return nil, err
		}
		blocks[i] = appMessageBlock
	}
	return &appmessage.GetBlocksResponseMessage{
		BlockHashes: x.BlockHashes,
		Blocks:      blocks,
		Error:       rpcErr,
	}, nil
}
