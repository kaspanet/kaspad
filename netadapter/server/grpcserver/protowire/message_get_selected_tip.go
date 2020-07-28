package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_GetSelectedTip) toWireMessage() (*wire.MsgGetSelectedTip, error) {
	return &wire.MsgGetSelectedTip{}, nil
}

func (x *KaspadMessage_GetSelectedTip) fromWireMessage(_ *wire.MsgGetSelectedTip) {
}
