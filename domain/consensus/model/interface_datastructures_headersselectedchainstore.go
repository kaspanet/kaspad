package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeadersSelectedChainStore represents a store of the headers selected chain
type HeadersSelectedChainStore interface {
	Store
	Stage(dbContext DBReader,
		chainChanges *externalapi.SelectedChainPath) error
	IsStaged() bool
	GetIndexByHash(dbContext DBReader, blockHash *externalapi.DomainHash) (uint64, error)
	GetHashByIndex(dbContext DBReader, index uint64) (*externalapi.DomainHash, error)
}
