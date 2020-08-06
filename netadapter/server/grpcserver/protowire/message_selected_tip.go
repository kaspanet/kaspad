package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_SelectedTip) toWireMessage() (domainmessage.Message, error) {
	hash, err := x.SelectedTip.SelectedTipHash.toWire()
	if err != nil {
		return nil, err
	}

	return &domainmessage.MsgSelectedTip{SelectedTipHash: hash}, nil
}

func (x *KaspadMessage_SelectedTip) fromWireMessage(msgSelectedTip *domainmessage.MsgSelectedTip) error {
	x.SelectedTip = &SelectedTipMessage{
		SelectedTipHash: wireHashToProto(msgSelectedTip.SelectedTipHash),
	}
	return nil
}
