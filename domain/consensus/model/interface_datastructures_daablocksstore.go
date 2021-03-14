package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAABlocksStore represents a store of ???
type DAABlocksStore interface {
	Store
	StageDAAScore(blockHash *externalapi.DomainHash, daaScore uint64)
	StageBlockDAAAddedBlocks(blockHash *externalapi.DomainHash, addedBlocks []*externalapi.DomainHash)
	IsAnythingStaged() bool
	DAAAddedBlocks(dbContext DBReader, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	DAAScore(dbContext DBReader, blockHash *externalapi.DomainHash) (uint64, error)
	Delete(blockHash *externalapi.DomainHash)
}
