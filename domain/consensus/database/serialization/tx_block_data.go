package serialization

import (
	"github.com/kaspanet/kaspad/domain/txindex"
)

// DomainOutpointToDbOutpoint converts DomainOutpoint to DbOutpoint
func TxBlockHashesToDbTxBlockHashes(txBlockHashes *txindex.TxBlockHashes) *DbTxBlockHashes {
	return &DbTxBlockHashes{
		AcceptingBlockHash: DomainHashToDbHash(txBlockHashes.acceptingBlockHash),
		MergeBlockHash:     DomainHashToDbHash(txBlockHashes.mergeBlockHash),
	}
}

// DbOutpointToDomainOutpoint converts DbOutpoint to DomainOutpoint
func DbTxBlockHashesToTxBlockHashes(dbTxBlockHashes *DbTxBlockHashes) (*txindex.TxBlockHashes, error) {
	acceptingBlockHash, err := DbHashToDomainHash(dbTxBlockHashes.AcceptingBlockHash)
	if err != nil {
		return nil, err
	}
	mergeBlockHash, err := DbHashToDomainHash(dbTxBlockHashes.AcceptingBlockHash)
	if err != nil {
		return nil, err
	}

	return &txindex.TxBlockHashes{
		AcceptingBlockHash: mergeBlockHash,
		MergeBlockHash:     acceptingBlockHash,
	}, nil
}
