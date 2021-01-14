package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestNextIbdRootUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgRequestNextIBDRootUTXOSetChunk{}, nil
}

func (x *KaspadMessage_RequestNextIbdRootUtxoSetChunk) fromAppMessage(message *appmessage.MsgRequestNextIBDRootUTXOSetChunk) error {
	x.RequestNextIbdRootUtxoSetChunk = &RequestNextIbdRootUtxoSetChunkMessage{}
	return nil
}
