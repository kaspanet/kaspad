package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/pkg/errors"
	"math"
	"math/big"
)

// DomainBlockHeaderToDbBlockHeader converts BlockHeader to DbBlockHeader
func DomainBlockHeaderToDbBlockHeader(domainBlockHeader externalapi.BlockHeader) *DbBlockHeader {
	return &DbBlockHeader{
		Version:              uint32(domainBlockHeader.Version()),
		Parents:              DomainParentsToDbParents(domainBlockHeader.Parents()),
		HashMerkleRoot:       DomainHashToDbHash(domainBlockHeader.HashMerkleRoot()),
		AcceptedIDMerkleRoot: DomainHashToDbHash(domainBlockHeader.AcceptedIDMerkleRoot()),
		UtxoCommitment:       DomainHashToDbHash(domainBlockHeader.UTXOCommitment()),
		TimeInMilliseconds:   domainBlockHeader.TimeInMilliseconds(),
		Bits:                 domainBlockHeader.Bits(),
		Nonce:                domainBlockHeader.Nonce(),
		DaaScore:             domainBlockHeader.DAAScore(),
		BlueScore:            domainBlockHeader.BlueScore(),
		BlueWork:             domainBlockHeader.BlueWork().Bytes(),
		PruningPoint:         DomainHashToDbHash(domainBlockHeader.PruningPoint()),
	}
}

// DbBlockHeaderToDomainBlockHeader converts DbBlockHeader to BlockHeader
func DbBlockHeaderToDomainBlockHeader(dbBlockHeader *DbBlockHeader) (externalapi.BlockHeader, error) {
	parents, err := DbParentsToDomainParents(dbBlockHeader.Parents)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot, err := DbHashToDomainHash(dbBlockHeader.HashMerkleRoot)
	if err != nil {
		return nil, err
	}
	acceptedIDMerkleRoot, err := DbHashToDomainHash(dbBlockHeader.AcceptedIDMerkleRoot)
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := DbHashToDomainHash(dbBlockHeader.UtxoCommitment)
	if err != nil {
		return nil, err
	}
	if dbBlockHeader.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid header version - bigger then uint16")
	}

	pruningPoint, err := DbHashToDomainHash(dbBlockHeader.PruningPoint)
	if err != nil {
		return nil, err
	}

	return blockheader.NewImmutableBlockHeader(
		uint16(dbBlockHeader.Version),
		parents,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		dbBlockHeader.TimeInMilliseconds,
		dbBlockHeader.Bits,
		dbBlockHeader.Nonce,
		dbBlockHeader.DaaScore,
		dbBlockHeader.BlueScore,
		new(big.Int).SetBytes(dbBlockHeader.BlueWork),
		pruningPoint,
	), nil
}
