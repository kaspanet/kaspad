package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DomainBlockHeaderToDbBlockHeader converts DomainBlockHeader to DbBlockHeader
func DomainBlockHeaderToDbBlockHeader(domainBlockHeader *externalapi.DomainBlockHeader) *DbBlockHeader {
	dbParentHashes := make([]*DbHash, len(domainBlockHeader.ParentHashes))
	for i, parentHash := range domainBlockHeader.ParentHashes {
		dbParentHashes[i] = DomainHashToDbHash(parentHash)
	}

	return &DbBlockHeader{
		Version:              domainBlockHeader.Version,
		ParentHashes:         dbParentHashes,
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
	parentHashes := make([]*externalapi.DomainHash, len(dbBlockHeader.ParentHashes))
	for i, dbParentHash := range dbBlockHeader.ParentHashes {
		var err error
		parentHashes[i], err = DbHashToDomainHash(dbParentHash)
		if err != nil {
			return nil, err
		}
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

	return &externalapi.DomainBlockHeader{
		Version:              dbBlockHeader.Version,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       *hashMerkleRoot,
		AcceptedIDMerkleRoot: *acceptedIDMerkleRoot,
		UTXOCommitment:       *utxoCommitment,
		TimeInMilliseconds:   dbBlockHeader.TimeInMilliseconds,
		Bits:                 dbBlockHeader.Bits,
		Nonce:                dbBlockHeader.Nonce,
	}, nil
}
