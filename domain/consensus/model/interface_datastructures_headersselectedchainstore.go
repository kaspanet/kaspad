package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeadersSelectedChainStore represents a store of the headers selected chain
type HeadersSelectedChainStore interface {
	Store
	Stage(dbContext DBReader, stagingArea *StagingArea, chainChanges *externalapi.SelectedChainPath) error
	IsStaged(stagingArea *StagingArea) bool
	GetIndexByHash(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (uint64, error)
	GetHashByIndex(dbContext DBReader, stagingArea *StagingArea, index uint64) (*externalapi.DomainHash, error)
}
