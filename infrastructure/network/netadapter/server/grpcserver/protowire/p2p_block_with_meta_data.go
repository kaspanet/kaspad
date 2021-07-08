package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"math/big"
)

func (x *KaspadMessage_BlockWithMetaData) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_BlockWithMetaData is nil")
	}

	msgBlock, err := x.BlockWithMetaData.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	daaWindow := make([]*appmessage.BlockWithMetaDataDAABlock, len(x.BlockWithMetaData.DaaWindow))
	for i, daaBlock := range x.BlockWithMetaData.DaaWindow {
		daaWindow[i], err = daaBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	ghostdagData := make([]*appmessage.BlockGHOSTDAGDataHashPair, len(x.BlockWithMetaData.GhostdagData))
	for i, pair := range x.BlockWithMetaData.GhostdagData {
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

	return &appmessage.MsgBlockWithMetaData{
		Block:        msgBlock,
		DAAScore:     x.BlockWithMetaData.DaaScore,
		DAAWindow:    daaWindow,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *KaspadMessage_BlockWithMetaData) fromAppMessage(msgBlockWithMetaData *appmessage.MsgBlockWithMetaData) error {
	x.BlockWithMetaData = &BlockWithMetaDataMessage{
		Block:        &BlockMessage{},
		DaaScore:     msgBlockWithMetaData.DAAScore,
		DaaWindow:    make([]*DaaBlock, len(msgBlockWithMetaData.DAAWindow)),
		GhostdagData: make([]*BlockGHOSTDAGDataHashPair, len(msgBlockWithMetaData.GHOSTDAGData)),
	}

	err := x.BlockWithMetaData.Block.fromAppMessage(msgBlockWithMetaData.Block)
	if err != nil {
		return err
	}

	for i, daaBlock := range msgBlockWithMetaData.DAAWindow {
		x.BlockWithMetaData.DaaWindow[i] = &DaaBlock{}
		err := x.BlockWithMetaData.DaaWindow[i].fromAppMessage(daaBlock)
		if err != nil {
			return err
		}
	}

	for i, pair := range msgBlockWithMetaData.GHOSTDAGData {
		x.BlockWithMetaData.GhostdagData[i] = &BlockGHOSTDAGDataHashPair{
			Hash:         domainHashToProto(pair.Hash),
			GhostdagData: &GhostdagData{},
		}

		x.BlockWithMetaData.GhostdagData[i].GhostdagData.fromAppMessage(pair.GHOSTDAGData)
	}

	return nil
}

func (x *DaaBlock) toAppMessage() (*appmessage.BlockWithMetaDataDAABlock, error) {
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

	return &appmessage.BlockWithMetaDataDAABlock{
		Header:       msgBlockHeader,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *DaaBlock) fromAppMessage(daaBlock *appmessage.BlockWithMetaDataDAABlock) error {
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
