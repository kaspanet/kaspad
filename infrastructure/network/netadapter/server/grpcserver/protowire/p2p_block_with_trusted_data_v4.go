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
		Block:               msgBlock,
		DAAWindowIndices:    x.BlockWithTrustedDataV4.DaaWindowIndices,
		GHOSTDAGDataIndices: x.BlockWithTrustedDataV4.GhostdagDataIndices,
	}, nil
}

func (x *KaspadMessage_BlockWithTrustedDataV4) fromAppMessage(msgBlockWithTrustedData *appmessage.MsgBlockWithTrustedDataV4) error {
	x.BlockWithTrustedDataV4 = &BlockWithTrustedDataV4Message{
		Block:               &BlockMessage{},
		DaaWindowIndices:    msgBlockWithTrustedData.DAAWindowIndices,
		GhostdagDataIndices: msgBlockWithTrustedData.GHOSTDAGDataIndices,
	}

	err := x.BlockWithTrustedDataV4.Block.fromAppMessage(msgBlockWithTrustedData.Block)
	if err != nil {
		return err
	}

	return nil
}

func (x *DaaBlockV4) toAppMessage() (*appmessage.TrustedDataDAAHeader, error) {
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

	return &appmessage.TrustedDataDAAHeader{
		Header:       msgBlockHeader,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *DaaBlockV4) fromAppMessage(daaBlock *appmessage.TrustedDataDAAHeader) error {
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
