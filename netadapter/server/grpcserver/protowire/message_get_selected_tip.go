package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_GetSelectedTip) toWireMessage() (wire.Message, error) {
	return &wire.MsgGetSelectedTip{}, nil
}

func (x *KaspadMessage_GetSelectedTip) fromWireMessage(_ *wire.MsgGetSelectedTip) error {
	return nil
}
