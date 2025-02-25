package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_AddArchivalBlocksRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_AddArchivalBlocksRequest is nil")
	}

	blocks := make([]*appmessage.ArchivalBlock, len(x.AddArchivalBlocksRequest.Blocks))
	for i, block := range x.AddArchivalBlocksRequest.Blocks {
		rpcBlock, err := block.Block.toAppMessage()
		if err != nil {
			return nil, err
		}

		blocks[i] = &appmessage.ArchivalBlock{
			Block: rpcBlock,
			Child: block.Child,
		}
	}

	return &appmessage.AddArchivalBlocksRequestMessage{
		Blocks: blocks,
	}, nil
}

func (x *KaspadMessage_AddArchivalBlocksRequest) fromAppMessage(message *appmessage.AddArchivalBlocksRequestMessage) error {
	blocks := make([]*ArchivalBlock, len(message.Blocks))
	for i, block := range message.Blocks {
		protoBlock := &ArchivalBlock{
			Child: block.Child,
		}

		if block.Block != nil {
			protoBlock.Block = &RpcBlock{}
			err := protoBlock.Block.fromAppMessage(block.Block)
			if err != nil {
				return err
			}
		}
		blocks[i] = protoBlock
	}

	x.AddArchivalBlocksRequest = &AddArchivalBlocksRequestMessage{
		Blocks: make([]*ArchivalBlock, len(message.Blocks)),
	}
	return nil
}

func (x *KaspadMessage_AddArchivalBlocksResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_AddArchivalBlocksResponse is nil")
	}
	return x.AddArchivalBlocksResponse.toAppMessage()
}

func (x *KaspadMessage_AddArchivalBlocksResponse) fromAppMessage(message *appmessage.AddArchivalBlocksResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}

	x.AddArchivalBlocksResponse = &AddArchivalBlocksResponseMessage{
		Error: err,
	}

	return nil
}

func (x *AddArchivalBlocksResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "AddArchivalBlocksResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	return &appmessage.GetPruningWindowRootsResponseMessage{
		Error: rpcErr,
	}, nil
}
