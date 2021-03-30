package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockHeaderStore represents a store of block headers
type BlockHeaderStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, blockHeader externalapi.BlockHeader)
	IsStaged(stagingArea *StagingArea) bool
	BlockHeader(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error)
	HasBlockHeader(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (bool, error)
	BlockHeaders(dbContext DBReader, stagingArea *StagingArea, blockHashes []*externalapi.DomainHash) ([]externalapi.BlockHeader, error)
	Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
	Count(stagingArea *StagingArea) uint64
}
