package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_RequestSelectedTip) toWireMessage() (wire.Message, error) {
	return &wire.MsgRequestSelectedTip{}, nil
}

func (x *KaspadMessage_RequestSelectedTip) fromWireMessage(_ *wire.MsgRequestSelectedTip) error {
	return nil
}
