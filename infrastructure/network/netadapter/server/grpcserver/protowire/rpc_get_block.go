package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockRequestMessage{
		Hash:                    x.GetBlockRequest.Hash,
		SubnetworkID:            x.GetBlockRequest.SubnetworkId,
		IncludeBlockHex:         x.GetBlockRequest.IncludeBlockHex,
		IncludeBlockVerboseData: x.GetBlockRequest.IncludeBlockVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlockRequest) fromAppMessage(message *appmessage.GetBlockRequestMessage) error {
	x.GetBlockRequest = &GetBlockRequestMessage{
		Hash:                    message.Hash,
		SubnetworkId:            message.SubnetworkID,
		IncludeBlockHex:         message.IncludeBlockHex,
		IncludeBlockVerboseData: message.IncludeBlockVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetBlockResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlockResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockResponse.Error.Message}
	}
	return &appmessage.GetBlockResponseMessage{
		BlockHex: x.GetBlockResponse.BlockHex,
		Error:    err,
	}, nil
}

func (x *KaspadMessage_GetBlockResponse) fromAppMessage(message *appmessage.GetBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlockResponse = &GetBlockResponseMessage{
		BlockHex: message.BlockHex,
		Error:    err,
	}
	return nil
}
