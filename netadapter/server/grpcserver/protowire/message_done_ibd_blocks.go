package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_DoneIBDBlocks) toWireMessage() (wire.Message, error) {
	return &wire.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneIBDBlocks) fromWireMessage(_ *wire.MsgDoneIBDBlocks) error {
	return nil
}
