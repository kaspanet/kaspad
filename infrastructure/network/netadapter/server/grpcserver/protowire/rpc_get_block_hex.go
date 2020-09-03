package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetBlockHexRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockHexRequestMessage{
		Hash:         x.GetBlockHexRequest.Hash,
		SubnetworkID: x.GetBlockHexRequest.SubnetworkId,
	}, nil
}

func (x *KaspadMessage_GetBlockHexRequest) fromAppMessage(message *appmessage.GetBlockHexRequestMessage) error {
	x.GetBlockHexRequest = &GetBlockHexRequestMessage{
		Hash:         message.Hash,
		SubnetworkId: message.SubnetworkID,
	}
	return nil
}

func (x *KaspadMessage_GetBlockHexResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlockHexResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockHexResponse.Error.Message}
	}
	return &appmessage.GetBlockHexResponseMessage{
		BlockHex: x.GetBlockHexResponse.BlockHex,
		Error:    err,
	}, nil
}

func (x *KaspadMessage_GetBlockHexResponse) fromAppMessage(message *appmessage.GetBlockHexResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlockHexResponse = &GetBlockHexResponseMessage{
		BlockHex: message.BlockHex,
		Error:    err,
	}
	return nil
}
