package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockWithTrustedDataV4) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_BlockWithTrustedDataV4 is nil")
	}

	msgBlock, err := x.BlockWithTrustedDataV4.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgBlockWithTrustedDataV4{
		Block:        msgBlock,
		DAAWindow:    x.BlockWithTrustedDataV4.DaaWindow,
		GHOSTDAGData: x.BlockWithTrustedDataV4.GhostdagData,
	}, nil
}

func (x *KaspadMessage_BlockWithTrustedDataV4) fromAppMessage(msgBlockWithTrustedData *appmessage.MsgBlockWithTrustedDataV4) error {
	x.BlockWithTrustedDataV4 = &BlockWithTrustedDataV4Message{
		Block:        &BlockMessage{},
		DaaWindow:    msgBlockWithTrustedData.DAAWindow,
		GhostdagData: msgBlockWithTrustedData.GHOSTDAGData,
	}

	err := x.BlockWithTrustedDataV4.Block.fromAppMessage(msgBlockWithTrustedData.Block)
	if err != nil {
		return err
	}

	return nil
}

func (x *DaaBlockV4) toAppMessage() (*appmessage.TrustedDataDataDAABlockV4, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "DaaBlockV4 is nil")
	}

	msgBlockHeader, err := x.Header.toAppMessage()
	if err != nil {
		return nil, err
	}

	ghostdagData, err := x.GhostdagData.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.TrustedDataDataDAABlockV4{
		Header:       msgBlockHeader,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *DaaBlockV4) fromAppMessage(daaBlock *appmessage.TrustedDataDataDAABlockV4) error {
	*x = DaaBlockV4{
		Header:       &BlockHeader{},
		GhostdagData: &GhostdagData{},
	}

	err := x.Header.fromAppMessage(daaBlock.Header)
	if err != nil {
		return err
	}

	x.GhostdagData.fromAppMessage(daaBlock.GHOSTDAGData)

	return nil
}
