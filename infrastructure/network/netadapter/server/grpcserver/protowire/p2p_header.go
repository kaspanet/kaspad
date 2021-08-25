package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"math"
	"math/big"
)

func (x *BlockHeader) toAppMessage() (*appmessage.MsgBlockHeader, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockHeaderMessage is nil")
	}
	parents, err := protoParentsToDomain(x.Parents)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot, err := x.HashMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}
	acceptedIDMerkleRoot, err := x.AcceptedIdMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := x.UtxoCommitment.toDomain()
	if err != nil {
		return nil, err
	}
	pruningPoint, err := x.PruningPoint.toDomain()
	if err != nil {
		return nil, err
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid block header version - bigger then uint16")
	}
	return &appmessage.MsgBlockHeader{
		Version:              uint16(x.Version),
		Parents:              parents,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            mstime.UnixMilliseconds(x.Timestamp),
		Bits:                 x.Bits,
		Nonce:                x.Nonce,
		DAAScore:             x.DaaScore,
		BlueScore:            x.BlueScore,
		BlueWork:             new(big.Int).SetBytes(x.BlueWork),
		PruningPoint:         pruningPoint,
	}, nil
}

func (x *BlockHeader) fromAppMessage(msgBlockHeader *appmessage.MsgBlockHeader) error {
	*x = BlockHeader{
		Version:              uint32(msgBlockHeader.Version),
		Parents:              domainParentsToProto(msgBlockHeader.Parents),
		HashMerkleRoot:       domainHashToProto(msgBlockHeader.HashMerkleRoot),
		AcceptedIdMerkleRoot: domainHashToProto(msgBlockHeader.AcceptedIDMerkleRoot),
		UtxoCommitment:       domainHashToProto(msgBlockHeader.UTXOCommitment),
		Timestamp:            msgBlockHeader.Timestamp.UnixMilliseconds(),
		Bits:                 msgBlockHeader.Bits,
		Nonce:                msgBlockHeader.Nonce,
		DaaScore:             msgBlockHeader.DAAScore,
		BlueScore:            msgBlockHeader.BlueScore,
		BlueWork:             msgBlockHeader.BlueWork.Bytes(),
		PruningPoint:         domainHashToProto(msgBlockHeader.PruningPoint),
	}
	return nil
}

func (x *BlockLevelParents) toDomain() (externalapi.BlockLevelParents, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "BlockLevelParents is nil")
	}
	domainBlockLevelParents := make(externalapi.BlockLevelParents, len(x.ParentHashes))
	for i, parentHash := range x.ParentHashes {
		var err error
		domainBlockLevelParents[i], err = externalapi.NewDomainHashFromByteSlice(parentHash.Bytes)
		if err != nil {
			return nil, err
		}
	}
	return domainBlockLevelParents, nil
}

func protoParentsToDomain(protoParents []*BlockLevelParents) ([]externalapi.BlockLevelParents, error) {
	domainParents := make([]externalapi.BlockLevelParents, len(protoParents))
	for i, protoBlockLevelParents := range protoParents {
		var err error
		domainParents[i], err = protoBlockLevelParents.toDomain()
		if err != nil {
			return nil, err
		}
	}
	return domainParents, nil
}

func domainBlockLevelParentsToProto(parentHashes externalapi.BlockLevelParents) *BlockLevelParents {
	protoParentHashes := make([]*Hash, len(parentHashes))
	for i, parentHash := range parentHashes {
		protoParentHashes[i] = &Hash{Bytes: parentHash.ByteSlice()}
	}
	return &BlockLevelParents{
		ParentHashes: protoParentHashes,
	}
}

func domainParentsToProto(parents []externalapi.BlockLevelParents) []*BlockLevelParents {
	protoParents := make([]*BlockLevelParents, len(parents))
	for i, hash := range parents {
		protoParents[i] = domainBlockLevelParentsToProto(hash)
	}
	return protoParents
}
