package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetBlocksRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlocksRequestMessage{
		LowHash:                 x.GetBlocksRequest.LowHash,
		IncludeBlockHexes:       x.GetBlocksRequest.IncludeBlockHexes,
		IncludeBlockVerboseData: x.GetBlocksRequest.IncludeBlockVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlocksRequest) fromAppMessage(message *appmessage.GetBlocksRequestMessage) error {
	x.GetBlocksRequest = &GetBlocksRequestMessage{
		LowHash:                 message.LowHash,
		IncludeBlockHexes:       message.IncludeBlockHexes,
		IncludeBlockVerboseData: message.IncludeBlockVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetBlocksResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlocksResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlocksResponse.Error.Message}
	}
	blockVerboseData := make([]*appmessage.BlockVerboseData, len(x.GetBlocksResponse.BlockVerboseData))
	for i, blockVerboseDatum := range x.GetBlocksResponse.BlockVerboseData {
		appBlockVerboseDatum, err := blockVerboseDatum.toAppMessage()
		if err != nil {
			return nil, err
		}
		blockVerboseData[i] = appBlockVerboseDatum
	}
	return &appmessage.GetBlocksResponseMessage{
		BlockHashes:      x.GetBlocksResponse.BlockHashes,
		BlockHexes:       x.GetBlocksResponse.BlockHexes,
		BlockVerboseData: blockVerboseData,
		Error:            err,
	}, nil
}

func (x *KaspadMessage_GetBlocksResponse) fromAppMessage(message *appmessage.GetBlocksResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
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
	x.GetBlocksResponse = &GetBlocksResponseMessage{
		BlockHashes:      message.BlockHashes,
		BlockHexes:       message.BlockHexes,
		BlockVerboseData: blockVerboseData,
		Error:            err,
	}
	return nil
}
