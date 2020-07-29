package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_IbdBlock) toWireMessage() (wire.Message, error) {
	msgBlock, err := x.IbdBlock.toWireMessage()
	if err != nil {
		return nil, err
	}
	return &wire.MsgIBDBlock{MsgBlock: *(msgBlock.(*wire.MsgBlock))}, nil
}

func (x *KaspadMessage_IbdBlock) fromWireMessage(msgIBDBlock *wire.MsgIBDBlock) error {
	x.IbdBlock = new(BlockMessage)
	return x.IbdBlock.fromWireMessage(&msgIBDBlock.MsgBlock)
}
