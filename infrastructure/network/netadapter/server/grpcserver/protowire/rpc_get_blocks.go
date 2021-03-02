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
		IncludeBlockVerboseData:       message.IncludeBlockVerboseData,
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
		IncludeBlockVerboseData:       x.IncludeBlockVerboseData,
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
		BlockHashes: message.BlockHashes,
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

func (x *GetBlocksResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlocksResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	// Return verbose data only if there's no error
	if rpcErr != nil && len(x.BlockVerboseData) != 0 {
		return nil, errors.New("GetBlocksResponseMessage contains both an error and a response")
	}
	blocksVerboseData := make([]*appmessage.BlockVerboseData, len(x.BlockVerboseData))
	for i, blockVerboseDatum := range x.BlockVerboseData {
		appBlockVerboseDatum, err := blockVerboseDatum.toAppMessage()
		if err != nil {
			return nil, err
		}
		blocksVerboseData[i] = appBlockVerboseDatum
	}
	return &appmessage.GetBlocksResponseMessage{
		BlockVerboseData: blocksVerboseData,
		Error:            rpcErr,
	}, nil
}
