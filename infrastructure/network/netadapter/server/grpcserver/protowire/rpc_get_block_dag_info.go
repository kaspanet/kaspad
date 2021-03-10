package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBlockDagInfoRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockDagInfoRequest is nil")
	}
	return &appmessage.GetBlockDAGInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetBlockDagInfoRequest) fromAppMessage(_ *appmessage.GetBlockDAGInfoRequestMessage) error {
	x.GetBlockDagInfoRequest = &GetBlockDagInfoRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetBlockDagInfoResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockDagInfoResponse is nil")
	}
	return x.GetBlockDagInfoResponse.toAppMessage()
}

func (x *GetBlockDagInfoResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockDagInfoResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	if rpcErr != nil && (x.NetworkName != "" || x.BlockCount != 0 || x.HeaderCount != 0 || len(x.TipHashes) != 0 || len(x.VirtualParentHashes) != 0 || x.Difficulty != 0 || x.PastMedianTime != 0 || x.PruningPointHash != "") {
		return nil, errors.New("GetBlockDagInfoResponseMessage contains both an error and a response")
	}
	return &appmessage.GetBlockDAGInfoResponseMessage{
		NetworkName:         x.NetworkName,
		BlockCount:          x.BlockCount,
		HeaderCount:         x.HeaderCount,
		TipHashes:           x.TipHashes,
		VirtualParentHashes: x.VirtualParentHashes,
		Difficulty:          x.Difficulty,
		PastMedianTime:      x.PastMedianTime,
		PruningPointHash:    x.PruningPointHash,
		Error:               rpcErr,
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
		PruningPointHash:    message.PruningPointHash,
		Error:               err,
	}
	return nil
}
