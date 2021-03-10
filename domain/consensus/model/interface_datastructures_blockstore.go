package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockStore represents a store of blocks
type BlockStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, block *externalapi.DomainBlock)
	IsStaged() bool
	Block(dbContext DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error)
	HasBlock(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
	Blocks(dbContext DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error)
	Delete(blockHash *externalapi.DomainHash)
	Count() uint64
	AllBlockHashesIterator(dbContext DBReader) (BlockIterator, error)
}
