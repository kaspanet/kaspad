package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockHeaderStore represents a store of block headers
type BlockHeaderStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, blockHeader externalapi.ImmutableBlockHeader)
	IsStaged() bool
	BlockHeader(dbContext DBReader, blockHash *externalapi.DomainHash) (externalapi.ImmutableBlockHeader, error)
	HasBlockHeader(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
	BlockHeaders(dbContext DBReader, blockHashes []*externalapi.DomainHash) ([]externalapi.ImmutableBlockHeader, error)
	Delete(blockHash *externalapi.DomainHash)
	Count() uint64
}
