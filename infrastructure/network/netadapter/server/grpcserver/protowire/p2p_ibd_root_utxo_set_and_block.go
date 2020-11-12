package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_IbdRootUTXOSetAndBlock) toAppMessage() (appmessage.Message, error) {
	msgBlock, err := x.IbdRootUTXOSetAndBlock.Block.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgIBDRootUTXOSetAndBlock{
		UTXOSet: x.IbdRootUTXOSetAndBlock.UtxoSet,
		Block:   msgBlock,
	}, nil
}

func (x *KaspadMessage_IbdRootUTXOSetAndBlock) fromAppMessage(msgIBDRootUTXOSetAndBlock *appmessage.MsgIBDRootUTXOSetAndBlock) error {
	x.IbdRootUTXOSetAndBlock = &IBDRootUTXOSetAndBlockMessage{}
	x.IbdRootUTXOSetAndBlock.UtxoSet = msgIBDRootUTXOSetAndBlock.UTXOSet
	x.IbdRootUTXOSetAndBlock.Block = &BlockMessage{}
	return x.IbdRootUTXOSetAndBlock.Block.fromAppMessage(msgIBDRootUTXOSetAndBlock.Block)
}
