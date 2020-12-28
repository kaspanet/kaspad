package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
)

// DomainBlockHeaderToDbBlockHeader converts ImmutableBlockHeader to DbBlockHeader
func DomainBlockHeaderToDbBlockHeader(domainBlockHeader externalapi.ImmutableBlockHeader) *DbBlockHeader {
	return &DbBlockHeader{
		Version:              domainBlockHeader.Version(),
		ParentHashes:         DomainHashesToDbHashes(domainBlockHeader.ParentHashes()),
		HashMerkleRoot:       DomainHashToDbHash(domainBlockHeader.HashMerkleRoot()),
		AcceptedIDMerkleRoot: DomainHashToDbHash(domainBlockHeader.AcceptedIDMerkleRoot()),
		UtxoCommitment:       DomainHashToDbHash(domainBlockHeader.UTXOCommitment()),
		TimeInMilliseconds:   domainBlockHeader.TimeInMilliseconds(),
		Bits:                 domainBlockHeader.Bits(),
		Nonce:                domainBlockHeader.Nonce(),
	}
}

// DbBlockHeaderToDomainBlockHeader converts DbBlockHeader to ImmutableBlockHeader
func DbBlockHeaderToDomainBlockHeader(dbBlockHeader *DbBlockHeader) (externalapi.ImmutableBlockHeader, error) {
	parentHashes, err := DbHashesToDomainHashes(dbBlockHeader.ParentHashes)
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

	return blockheader.NewImmutableBlockHeader(
		dbBlockHeader.Version,
		parentHashes,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		dbBlockHeader.TimeInMilliseconds,
		dbBlockHeader.Bits,
		dbBlockHeader.Nonce,
	), nil
}
