package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestNextPruningPointUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestNextPruningPointUTXOSetChunk{}, nil
}

func (x *KaspadMessage_RequestNextPruningPointUtxoSetChunk) fromAppMessage(_ *appmessage.MsgRequestNextPruningPointUTXOSetChunk) error {
	x.RequestNextPruningPointUtxoSetChunk = &RequestNextPruningPointUtxoSetChunkMessage{}
	return nil
}
