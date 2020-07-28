package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_Verack) toWireMessage() (*wire.MsgVerAck, error) {
	return &wire.MsgVerAck{}, nil
}

func (x *KaspadMessage_Verack) fromWireMessage(_ *wire.MsgVerAck) {
}
