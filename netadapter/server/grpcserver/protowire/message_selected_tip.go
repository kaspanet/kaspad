package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_SelectedTip) toWireMessage() (wire.Message, error) {
	hash, err := x.SelectedTip.SelectedTipHash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgSelectedTip{SelectedTipHash: hash}, nil
}

func (x *KaspadMessage_SelectedTip) fromWireMessage(msgSelectedTip *wire.MsgSelectedTip) error {
	x.SelectedTip = &SelectedTipMessage{
		SelectedTipHash: wireHashToProto(msgSelectedTip.SelectedTipHash),
	}
	return nil
}
