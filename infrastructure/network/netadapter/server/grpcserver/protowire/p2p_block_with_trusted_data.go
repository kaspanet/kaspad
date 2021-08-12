package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"math/big"
)

func (x *KaspadMessage_BlockWithTrustedData) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_BlockWithTrustedData is nil")
	}

	msgBlock, err := x.BlockWithTrustedData.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	daaWindow := make([]*appmessage.TrustedDataDataDAABlock, len(x.BlockWithTrustedData.DaaWindow))
	for i, daaBlock := range x.BlockWithTrustedData.DaaWindow {
		daaWindow[i], err = daaBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	ghostdagData := make([]*appmessage.BlockGHOSTDAGDataHashPair, len(x.BlockWithTrustedData.GhostdagData))
	for i, pair := range x.BlockWithTrustedData.GhostdagData {
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

	return &appmessage.MsgBlockWithTrustedData{
		Block:        msgBlock,
		DAAScore:     x.BlockWithTrustedData.DaaScore,
		DAAWindow:    daaWindow,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *KaspadMessage_BlockWithTrustedData) fromAppMessage(msgBlockWithTrustedData *appmessage.MsgBlockWithTrustedData) error {
	x.BlockWithTrustedData = &BlockWithTrustedDataMessage{
		Block:        &BlockMessage{},
		DaaScore:     msgBlockWithTrustedData.DAAScore,
		DaaWindow:    make([]*DaaBlock, len(msgBlockWithTrustedData.DAAWindow)),
		GhostdagData: make([]*BlockGhostdagDataHashPair, len(msgBlockWithTrustedData.GHOSTDAGData)),
	}

	err := x.BlockWithTrustedData.Block.fromAppMessage(msgBlockWithTrustedData.Block)
	if err != nil {
		return err
	}

	for i, daaBlock := range msgBlockWithTrustedData.DAAWindow {
		x.BlockWithTrustedData.DaaWindow[i] = &DaaBlock{}
		err := x.BlockWithTrustedData.DaaWindow[i].fromAppMessage(daaBlock)
		if err != nil {
			return err
		}
	}

	for i, pair := range msgBlockWithTrustedData.GHOSTDAGData {
		x.BlockWithTrustedData.GhostdagData[i] = &BlockGhostdagDataHashPair{
			Hash:         domainHashToProto(pair.Hash),
			GhostdagData: &GhostdagData{},
		}

		x.BlockWithTrustedData.GhostdagData[i].GhostdagData.fromAppMessage(pair.GHOSTDAGData)
	}

	return nil
}

func (x *DaaBlock) toAppMessage() (*appmessage.TrustedDataDataDAABlock, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "DaaBlock is nil")
	}

	msgBlockHeader, err := x.Header.toAppMessage()
	if err != nil {
		return nil, err
	}

	ghostdagData, err := x.GhostdagData.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.TrustedDataDataDAABlock{
		Header:       msgBlockHeader,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *DaaBlock) fromAppMessage(daaBlock *appmessage.TrustedDataDataDAABlock) error {
	*x = DaaBlock{
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

func (x *GhostdagData) toAppMessage() (*appmessage.BlockGHOSTDAGData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GhostdagData is nil")
	}

	selectedParent, err := x.SelectedParent.toDomain()
	if err != nil {
		return nil, err
	}

	mergeSetBlues, err := protoHashesToDomain(x.MergeSetBlues)
	if err != nil {
		return nil, err
	}

	mergeSetReds, err := protoHashesToDomain(x.MergeSetReds)
	if err != nil {
		return nil, err
	}

	bluesAnticoneSizes := make([]*appmessage.BluesAnticoneSizes, len(x.BluesAnticoneSizes))
	for i, protoBluesAnticoneSizes := range x.BluesAnticoneSizes {
		blueHash, err := protoBluesAnticoneSizes.BlueHash.toDomain()
		if err != nil {
			return nil, err
		}

		if protoBluesAnticoneSizes.AnticoneSize > maxKType() {
			return nil, errors.Errorf("anticone size %d is greater than max k type %d", protoBluesAnticoneSizes.AnticoneSize, maxKType())
		}

		bluesAnticoneSizes[i] = &appmessage.BluesAnticoneSizes{
			BlueHash:     blueHash,
			AnticoneSize: externalapi.KType(protoBluesAnticoneSizes.AnticoneSize),
		}
	}

	blueWork := big.NewInt(0).SetBytes(x.BlueWork)
	return &appmessage.BlockGHOSTDAGData{
		BlueScore:          x.BlueScore,
		BlueWork:           blueWork,
		SelectedParent:     selectedParent,
		MergeSetBlues:      mergeSetBlues,
		MergeSetReds:       mergeSetReds,
		BluesAnticoneSizes: bluesAnticoneSizes,
	}, nil
}

func (x *GhostdagData) fromAppMessage(ghostdagData *appmessage.BlockGHOSTDAGData) {
	protoBluesAnticoneSizes := make([]*BluesAnticoneSizes, 0, len(ghostdagData.BluesAnticoneSizes))
	for _, pair := range ghostdagData.BluesAnticoneSizes {
		protoBluesAnticoneSizes = append(protoBluesAnticoneSizes, &BluesAnticoneSizes{
			BlueHash:     domainHashToProto(pair.BlueHash),
			AnticoneSize: uint32(pair.AnticoneSize),
		})
	}
	*x = GhostdagData{
		BlueScore:          ghostdagData.BlueScore,
		BlueWork:           ghostdagData.BlueWork.Bytes(),
		SelectedParent:     domainHashToProto(ghostdagData.SelectedParent),
		MergeSetBlues:      domainHashesToProto(ghostdagData.MergeSetBlues),
		MergeSetReds:       domainHashesToProto(ghostdagData.MergeSetReds),
		BluesAnticoneSizes: protoBluesAnticoneSizes,
	}
}

func maxKType() uint32 {
	zero := externalapi.KType(0)
	max := zero - 1
	return uint32(max)
}
