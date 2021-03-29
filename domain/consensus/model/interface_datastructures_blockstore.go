package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockStore represents a store of blocks
type BlockStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, block *externalapi.DomainBlock)
	IsStaged(stagingArea *StagingArea) bool
	Block(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error)
	HasBlock(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (bool, error)
	Blocks(dbContext DBReader, stagingArea *StagingArea, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error)
	Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
	Count(stagingArea *StagingArea) uint64
	AllBlockHashesIterator(dbContext DBReader) (BlockIterator, error)
}
