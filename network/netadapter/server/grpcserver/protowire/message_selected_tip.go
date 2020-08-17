package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_SelectedTip) toAppMessage() (appmessage.Message, error) {
	hash, err := x.SelectedTip.SelectedTipHash.toWire()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgSelectedTip{SelectedTipHash: hash}, nil
}

func (x *KaspadMessage_SelectedTip) fromAppMessage(msgSelectedTip *appmessage.MsgSelectedTip) error {
	x.SelectedTip = &SelectedTipMessage{
		SelectedTipHash: wireHashToProto(msgSelectedTip.SelectedTipHash),
	}
	return nil
}
