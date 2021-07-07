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
	domainBlock := appmessage.MsgBlockToDomainBlock(msgBlock)

	daaWindow := make([]*externalapi.BlockWithMetaDataDAABlock, len(x.BlockWithMetaData.DaaWindow))
	for i, daaBlock := range x.BlockWithMetaData.DaaWindow {
		daaWindow[i], err = daaBlock.toDomain()
		if err != nil {
			return nil, err
		}
	}

	ghostdagData := make([]*externalapi.BlockGHOSTDAGDataHashPair, len(x.BlockWithMetaData.GhostdagData))
	for i, pair := range x.BlockWithMetaData.GhostdagData {
		hash, err := pair.Hash.toDomain()
		if err != nil {
			return nil, err
		}

		data, err := pair.GhostdagData.toDomain()
		if err != nil {
			return nil, err
		}

		ghostdagData[i] = &externalapi.BlockGHOSTDAGDataHashPair{
			Hash:         hash,
			GHOSTDAGData: data,
		}
	}

	return appmessage.NewMsgBlockWithMetaData(&externalapi.BlockWithMetaData{
		Block:        domainBlock,
		DAAScore:     x.BlockWithMetaData.DaaScore,
		DAAWindow:    daaWindow,
		GHOSTDAGData: ghostdagData,
	}), nil
}

func (x *KaspadMessage_BlockWithMetaData) fromAppMessage(msgBlockWithMetaData *appmessage.MsgBlockWithMetaData) error {
	msgBlock := appmessage.DomainBlockToMsgBlock(msgBlockWithMetaData.Block)
	x.BlockWithMetaData = &BlockWithMetaDataMessage{
		Block:        &BlockMessage{},
		DaaScore:     msgBlockWithMetaData.DAAScore,
		DaaWindow:    make([]*DaaBlock, len(msgBlockWithMetaData.DAAWindow)),
		GhostdagData: make([]*BlockGHOSTDAGDataHashPair, len(msgBlockWithMetaData.GHOSTDAGData)),
	}

	err := x.BlockWithMetaData.Block.fromAppMessage(msgBlock)
	if err != nil {
		return err
	}

	for i, daaBlock := range msgBlockWithMetaData.DAAWindow {
		x.BlockWithMetaData.DaaWindow[i] = &DaaBlock{}
		err := x.BlockWithMetaData.DaaWindow[i].fromDomain(daaBlock)
		if err != nil {
			return err
		}
	}

	for i, pair := range msgBlockWithMetaData.GHOSTDAGData {
		x.BlockWithMetaData.GhostdagData[i] = &BlockGHOSTDAGDataHashPair{
			Hash:         domainHashToProto(pair.Hash),
			GhostdagData: &GhostdagData{},
		}

		x.BlockWithMetaData.GhostdagData[i].GhostdagData.fromDomain(pair.GHOSTDAGData)
	}

	return nil
}

func (x *DaaBlock) toDomain() (*externalapi.BlockWithMetaDataDAABlock, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "DaaBlock is nil")
	}

	msgBlockHeader, err := x.Header.toAppMessage()
	if err != nil {
		return nil, err
	}

	header := appmessage.BlockHeaderToDomainBlockHeader(msgBlockHeader)

	ghostdagData, err := x.GhostdagData.toDomain()
	if err != nil {
		return nil, err
	}

	return &externalapi.BlockWithMetaDataDAABlock{
		Header:       header,
		GHOSTDAGData: ghostdagData,
	}, nil
}

func (x *DaaBlock) fromDomain(daaBlock *externalapi.BlockWithMetaDataDAABlock) error {
	*x = DaaBlock{
		Header:       &BlockHeader{},
		GhostdagData: &GhostdagData{},
	}

	err := x.Header.fromAppMessage(appmessage.DomainBlockHeaderToBlockHeader(daaBlock.Header))
	if err != nil {
		return err
	}

	x.GhostdagData.fromDomain(daaBlock.GHOSTDAGData)

	return nil
}

func (x *GhostdagData) toDomain() (*externalapi.BlockGHOSTDAGData, error) {
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

	bluesAnticoneSizes := make(map[externalapi.DomainHash]externalapi.KType, len(x.BluesAnticoneSizes))
	for _, protoBluesAnticoneSizes := range x.BluesAnticoneSizes {
		blueHash, err := protoBluesAnticoneSizes.BlueHash.toDomain()
		if err != nil {
			return nil, err
		}

		if protoBluesAnticoneSizes.AnticoneSize > maxKType() {
			bluesAnticoneSizes[*blueHash] = externalapi.KType(protoBluesAnticoneSizes.AnticoneSize)
		}
	}

	blueWork := big.NewInt(0).SetBytes(x.BlueWork)

	return externalapi.NewBlockGHOSTDAGData(
		x.BlueScore,
		blueWork,
		selectedParent,
		mergeSetBlues,
		mergeSetReds,
		bluesAnticoneSizes,
	), nil
}

func (x *GhostdagData) fromDomain(ghostdagData *externalapi.BlockGHOSTDAGData) {
	protoBluesAnticoneSizes := make([]*BluesAnticoneSizes, 0, len(ghostdagData.BluesAnticoneSizes()))
	for blueHash, anitconeSize := range ghostdagData.BluesAnticoneSizes() {
		protoBluesAnticoneSizes = append(protoBluesAnticoneSizes, &BluesAnticoneSizes{
			BlueHash:     domainHashToProto(&blueHash),
			AnticoneSize: uint32(anitconeSize),
		})
	}
	*x = GhostdagData{
		BlueScore:          ghostdagData.BlueScore(),
		BlueWork:           ghostdagData.BlueWork().Bytes(),
		SelectedParent:     domainHashToProto(ghostdagData.SelectedParent()),
		MergeSetBlues:      domainHashesToProto(ghostdagData.MergeSetBlues()),
		MergeSetReds:       domainHashesToProto(ghostdagData.MergeSetReds()),
		BluesAnticoneSizes: protoBluesAnticoneSizes,
	}
}

func maxKType() uint32 {
	zero := externalapi.KType(0)
	max := zero - 1
	return uint32(max)
}
