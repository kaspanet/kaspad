package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_DonePruningPointUtxoSetChunks) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgDonePruningPointUTXOSetChunks{}, nil
}

func (x *KaspadMessage_DonePruningPointUtxoSetChunks) fromAppMessage(_ *appmessage.MsgDonePruningPointUTXOSetChunks) error {
	x.DonePruningPointUtxoSetChunks = &DonePruningPointUtxoSetChunksMessage{}
	return nil
}
