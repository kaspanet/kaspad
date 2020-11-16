package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockHeaderStore represents a store of block headers
type BlockHeaderStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, blockHeader *externalapi.DomainBlockHeader) error
	IsStaged() bool
	BlockHeader(dbContext DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error)
	HasBlockHeader(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
	BlockHeaders(dbContext DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlockHeader, error)
	Delete(blockHash *externalapi.DomainHash)
	Count(dbContext DBReader) (uint64, error)
}
