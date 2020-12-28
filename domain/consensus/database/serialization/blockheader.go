package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// DomainBlockHeaderToDbBlockHeader converts DomainBlockHeader to DbBlockHeader
func DomainBlockHeaderToDbBlockHeader(domainBlockHeader *externalapi.DomainBlockHeader) *DbBlockHeader {
	return &DbBlockHeader{
		Version:              uint32(domainBlockHeader.Version),
		ParentHashes:         DomainHashesToDbHashes(domainBlockHeader.ParentHashes),
		HashMerkleRoot:       DomainHashToDbHash(&domainBlockHeader.HashMerkleRoot),
		AcceptedIDMerkleRoot: DomainHashToDbHash(&domainBlockHeader.AcceptedIDMerkleRoot),
		UtxoCommitment:       DomainHashToDbHash(&domainBlockHeader.UTXOCommitment),
		TimeInMilliseconds:   domainBlockHeader.TimeInMilliseconds,
		Bits:                 domainBlockHeader.Bits,
		Nonce:                domainBlockHeader.Nonce,
	}
}

// DbBlockHeaderToDomainBlockHeader converts DbBlockHeader to DomainBlockHeader
func DbBlockHeaderToDomainBlockHeader(dbBlockHeader *DbBlockHeader) (*externalapi.DomainBlockHeader, error) {
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
	if dbBlockHeader.Version > 0xffff {
		return nil, errors.Errorf("Invalid version size - bigger then uint16")
	}
	return &externalapi.DomainBlockHeader{
		Version:              uint16(dbBlockHeader.Version),
		ParentHashes:         parentHashes,
		HashMerkleRoot:       *hashMerkleRoot,
		AcceptedIDMerkleRoot: *acceptedIDMerkleRoot,
		UTXOCommitment:       *utxoCommitment,
		TimeInMilliseconds:   dbBlockHeader.TimeInMilliseconds,
		Bits:                 dbBlockHeader.Bits,
		Nonce:                dbBlockHeader.Nonce,
	}, nil
}
