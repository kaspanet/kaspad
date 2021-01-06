package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage{
		StartHash: x.GetVirtualSelectedParentChainFromBlockRequest.StartHash,
	}, nil
}

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest) fromAppMessage(message *appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage) error {
	x.GetVirtualSelectedParentChainFromBlockRequest = &GetVirtualSelectedParentChainFromBlockRequestMessage{
		StartHash: message.StartHash,
	}
	return nil
}

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetVirtualSelectedParentChainFromBlockResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetVirtualSelectedParentChainFromBlockResponse.Error.Message}
	}
	addedChainBlocks := make([]*appmessage.ChainBlock, len(x.GetVirtualSelectedParentChainFromBlockResponse.AddedChainBlocks))
	for i, addedChainBlock := range x.GetVirtualSelectedParentChainFromBlockResponse.AddedChainBlocks {
		appAddedChainBlock, err := addedChainBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
		addedChainBlocks[i] = appAddedChainBlock
	}
	return &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{
		RemovedChainBlockHashes: x.GetVirtualSelectedParentChainFromBlockResponse.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
		Error:                   err,
	}, nil
}

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockResponse) fromAppMessage(message *appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	addedChainBlocks := make([]*ChainBlock, len(message.AddedChainBlocks))
	for i, addedChainBlock := range message.AddedChainBlocks {
		protoAddedChainBlock := &ChainBlock{}
		err := protoAddedChainBlock.fromAppMessage(addedChainBlock)
		if err != nil {
			return err
		}
		addedChainBlocks[i] = protoAddedChainBlock
	}
	x.GetVirtualSelectedParentChainFromBlockResponse = &GetVirtualSelectedParentChainFromBlockResponseMessage{
		RemovedChainBlockHashes: message.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
		Error:                   err,
	}
	return nil
}
