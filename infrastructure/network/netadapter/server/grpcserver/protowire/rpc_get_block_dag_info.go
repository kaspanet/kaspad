package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetBlockDagInfoRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockDAGInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetBlockDagInfoRequest) fromAppMessage(_ *appmessage.GetBlockDAGInfoRequestMessage) error {
	x.GetBlockDagInfoRequest = &GetBlockDagInfoRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetBlockDagInfoResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlockDagInfoResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockDagInfoResponse.Error.Message}
	}
	return &appmessage.GetBlockDAGInfoResponseMessage{
		NetworkName:         x.GetBlockDagInfoResponse.NetworkName,
		BlockCount:          x.GetBlockDagInfoResponse.BlockCount,
		HeaderCount:         x.GetBlockDagInfoResponse.HeaderCount,
		TipHashes:           x.GetBlockDagInfoResponse.TipHashes,
		VirtualParentHashes: x.GetBlockDagInfoResponse.VirtualParentHashes,
		Difficulty:          x.GetBlockDagInfoResponse.Difficulty,
		PastMedianTime:      x.GetBlockDagInfoResponse.PastMedianTime,
		Error:               err,
	}, nil
}

func (x *KaspadMessage_GetBlockDagInfoResponse) fromAppMessage(message *appmessage.GetBlockDAGInfoResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlockDagInfoResponse = &GetBlockDagInfoResponseMessage{
		NetworkName:         message.NetworkName,
		BlockCount:          message.BlockCount,
		HeaderCount:         message.HeaderCount,
		TipHashes:           message.TipHashes,
		VirtualParentHashes: message.VirtualParentHashes,
		Difficulty:          message.Difficulty,
		PastMedianTime:      message.PastMedianTime,
		Error:               err,
	}
	return nil
}
