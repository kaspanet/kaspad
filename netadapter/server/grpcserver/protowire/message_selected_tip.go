package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_SelectedTip_) toWireMessage() (*wire.MsgSelectedTip, error) {
	hash, err := x.SelectedTip_.SelectedTipHash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgSelectedTip{SelectedTipHash: hash}, nil
}

func (x *KaspadMessage_SelectedTip_) fromWireMessage(msgSelectedTip *wire.MsgSelectedTip) {
	x.SelectedTip_ = &SelectedTipMessage{
		SelectedTipHash: wireHashToProto(msgSelectedTip.SelectedTipHash),
	}
}
