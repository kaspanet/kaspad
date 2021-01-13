package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_IbdRootUtxoSetChunk) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgIBDRootUTXOSetChunk{
		Chunk: x.IbdRootUtxoSetChunk.Chunk,
	}, nil
}

func (x *KaspadMessage_IbdRootUtxoSetChunk) fromAppMessage(message *appmessage.MsgIBDRootUTXOSetChunk) error {
	x.IbdRootUtxoSetChunk = &IbdRootUtxoSetChunkMessage{
		Chunk: message.Chunk,
	}
	return nil
}
