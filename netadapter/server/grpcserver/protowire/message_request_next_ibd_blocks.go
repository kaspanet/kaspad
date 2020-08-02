package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_RequestNextIBDBlocks) toWireMessage() (wire.Message, error) {
	return &wire.MsgRequestNextIBDBlocks{}, nil
}

func (x *KaspadMessage_RequestNextIBDBlocks) fromWireMessage(_ *wire.MsgRequestNextIBDBlocks) error {
	return nil
}
