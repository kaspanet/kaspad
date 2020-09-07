package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetChainFromBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetChainFromBlockRequestMessage{
		StartHash:               x.GetChainFromBlockRequest.StartHash,
		IncludeBlockVerboseData: x.GetChainFromBlockRequest.IncludeBlockVerboseData,
	}, nil
}

func (x *KaspadMessage_GetChainFromBlockRequest) fromAppMessage(message *appmessage.GetChainFromBlockRequestMessage) error {
	x.GetChainFromBlockRequest = &GetChainFromBlockRequestMessage{
		StartHash:               message.StartHash,
		IncludeBlockVerboseData: message.IncludeBlockVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetChainFromBlockResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetChainFromBlockResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetChainFromBlockResponse.Error.Message}
	}
	addedChainBlocks := make([]*appmessage.ChainBlock, len(x.GetChainFromBlockResponse.AddedChainBlocks))
	for i, addedChainBlock := range x.GetChainFromBlockResponse.AddedChainBlocks {
		appAddedChainBlock, err := addedChainBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
		addedChainBlocks[i] = appAddedChainBlock
	}
	blockVerboseData := make([]*appmessage.BlockVerboseData, len(x.GetChainFromBlockResponse.BlockVerboseData))
	for i, blockVerboseDatum := range x.GetChainFromBlockResponse.BlockVerboseData {
		appBlockVerboseDatum, err := blockVerboseDatum.toAppMessage()
		if err != nil {
			return nil, err
		}
		blockVerboseData[i] = appBlockVerboseDatum
	}
	return &appmessage.GetChainFromBlockResponseMessage{
		RemovedChainBlockHashes: x.GetChainFromBlockResponse.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
		BlockVerboseData:        blockVerboseData,
		Error:                   err,
	}, nil
}

func (x *KaspadMessage_GetChainFromBlockResponse) fromAppMessage(message *appmessage.GetChainFromBlockResponseMessage) error {
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
	blockVerboseData := make([]*BlockVerboseData, len(message.BlockVerboseData))
	for i, blockVerboseDatum := range message.BlockVerboseData {
		protoBlockVerboseDatum := &BlockVerboseData{}
		err := protoBlockVerboseDatum.fromAppMessage(blockVerboseDatum)
		if err != nil {
			return err
		}
		blockVerboseData[i] = protoBlockVerboseDatum
	}
	x.GetChainFromBlockResponse = &GetChainFromBlockResponseMessage{
		RemovedChainBlockHashes: message.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
		BlockVerboseData:        blockVerboseData,
		Error:                   err,
	}
	return nil
}
