package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_DoneIbdRootUtxoSetChunks) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgDoneIBDRootUTXOSetChunks{}, nil
}

func (x *KaspadMessage_DoneIbdRootUtxoSetChunks) fromAppMessage(_ *appmessage.MsgDoneIBDRootUTXOSetChunks) error {
	x.DoneIbdRootUtxoSetChunks = &DoneIbdRootUtxoSetChunksMessage{}
	return nil
}
