package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_SelectedTip) toAppMessage() (appmessage.Message, error) {
	hash, err := x.SelectedTip.SelectedTipHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgSelectedTip{SelectedTipHash: hash}, nil
}

func (x *KaspadMessage_SelectedTip) fromAppMessage(msgSelectedTip *appmessage.MsgSelectedTip) error {
	x.SelectedTip = &SelectedTipMessage{
		SelectedTipHash: domainHashToProto(msgSelectedTip.SelectedTipHash),
	}
	return nil
}
