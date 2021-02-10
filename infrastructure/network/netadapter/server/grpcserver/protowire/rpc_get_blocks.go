package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetBlocksRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlocksRequestMessage{
		LowHash:                       x.GetBlocksRequest.LowHash,
		IncludeBlockVerboseData:       x.GetBlocksRequest.IncludeBlockVerboseData,
		IncludeTransactionVerboseData: x.GetBlocksRequest.IncludeTransactionVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlocksRequest) fromAppMessage(message *appmessage.GetBlocksRequestMessage) error {
	x.GetBlocksRequest = &GetBlocksRequestMessage{
		LowHash:                       message.LowHash,
		IncludeBlockVerboseData:       message.IncludeBlockVerboseData,
		IncludeTransactionVerboseData: message.IncludeTransactionVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetBlocksResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlocksResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlocksResponse.Error.Message}
	}
	appMessage := &appmessage.GetBlocksResponseMessage{
		BlockHashes: x.GetBlocksResponse.BlockHashes,
		NextLowHash: x.GetBlocksResponse.NextLowHash,
		Error:       err,
	}
	if x.GetBlocksResponse.BlockVerboseData != nil {
		appMessage.BlockVerboseData = make([]*appmessage.BlockVerboseData, len(x.GetBlocksResponse.BlockVerboseData))
		for i, blockVerboseDatum := range x.GetBlocksResponse.BlockVerboseData {
			appBlockVerboseDatum, err := blockVerboseDatum.toAppMessage()
			if err != nil {
				return nil, err
			}
			appMessage.BlockVerboseData[i] = appBlockVerboseDatum
		}
	}

	return appMessage, nil
}

func (x *KaspadMessage_GetBlocksResponse) fromAppMessage(message *appmessage.GetBlocksResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlocksResponse = &GetBlocksResponseMessage{
		BlockHashes: message.BlockHashes,
		NextLowHash: message.NextLowHash,
		Error:       err,
	}
	if message.BlockVerboseData != nil {
		x.GetBlocksResponse.BlockVerboseData = make([]*BlockVerboseData, len(message.BlockVerboseData))
		for i, blockVerboseDatum := range message.BlockVerboseData {
			protoBlockVerboseDatum := &BlockVerboseData{}
			err := protoBlockVerboseDatum.fromAppMessage(blockVerboseDatum)
			if err != nil {
				return err
			}
			x.GetBlocksResponse.BlockVerboseData[i] = protoBlockVerboseDatum
		}
	}
	return nil
}
