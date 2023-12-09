package protowire

import (
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_RequestNextPruningPointUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestNextPruningPointUtxoSetChunk is nil")
	}
	return &appmessage.MsgRequestNextPruningPointUTXOSetChunk{}, nil
}

func (x *KaspadMessage_RequestNextPruningPointUtxoSetChunk) fromAppMessage(_ *appmessage.MsgRequestNextPruningPointUTXOSetChunk) error {
	x.RequestNextPruningPointUtxoSetChunk = &RequestNextPruningPointUtxoSetChunkMessage{}
	return nil
}
