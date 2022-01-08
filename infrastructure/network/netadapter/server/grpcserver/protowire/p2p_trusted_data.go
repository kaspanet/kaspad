package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_TrustedDataMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_TrustedDataMessage is nil")
	}

	daaWindow := make([]*appmessage.TrustedDataDataDAABlockV4, len(x.TrustedDataMessage.DaaWindow))
	for i, daaBlock := range x.TrustedDataMessage.DaaWindow {
		var err error
		daaWindow[i], err = daaBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	ghostdagData := make([]*appmessage.BlockGHOSTDAGDataHashPair, len(x.TrustedDataMessage.GhostdagData))
	for i, pair := range x.TrustedDataMessage.GhostdagData {
		hash, err := pair.Hash.toDomain()
		if err != nil {
			return nil, err
		}

		data, err := pair.GhostdagData.toAppMessage()
		if err != nil {
			return nil, err
		}

		ghostdagData[i] = &appmessage.BlockGHOSTDAGDataHashPair{
			Hash:         hash,
			GHOSTDAGData: data,
		}
	}

	return &appmessage.MsgTrustedData{
		DAAWindow:    daaWindow,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *KaspadMessage_TrustedDataMessage) fromAppMessage(msgTrustedData *appmessage.MsgTrustedData) error {
	x.TrustedDataMessage = &TrustedDataMessage{
		DaaWindow:    make([]*DaaBlockV4, len(msgTrustedData.DAAWindow)),
		GhostdagData: make([]*BlockGhostdagDataHashPair, len(msgTrustedData.GHOSTDAGData)),
	}

	for i, daaBlock := range msgTrustedData.DAAWindow {
		x.TrustedDataMessage.DaaWindow[i] = &DaaBlockV4{}
		err := x.TrustedDataMessage.DaaWindow[i].fromAppMessage(daaBlock)
		if err != nil {
			return err
		}
	}

	for i, pair := range msgTrustedData.GHOSTDAGData {
		x.TrustedDataMessage.GhostdagData[i] = &BlockGhostdagDataHashPair{
			Hash:         domainHashToProto(pair.Hash),
			GhostdagData: &GhostdagData{},
		}

		x.TrustedDataMessage.GhostdagData[i].GhostdagData.fromAppMessage(pair.GHOSTDAGData)
	}

	return nil
}
