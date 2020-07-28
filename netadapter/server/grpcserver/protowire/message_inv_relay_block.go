package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_InvRelayBlock) toWireMessage() (*wire.MsgInvRelayBlock, error) {
	hash, err := x.InvRelayBlock.Hash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgInvRelayBlock{Hash: hash}, nil
}

func (x *KaspadMessage_InvRelayBlock) fromWireMessage(msgInvRelayBlock *wire.MsgInvRelayBlock) error {
	x.InvRelayBlock = &InvRelayBlockMessage{
		Hash: wireHashToProto(msgInvRelayBlock.Hash),
	}
	return nil
}
