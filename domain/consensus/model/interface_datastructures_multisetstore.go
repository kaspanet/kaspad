package model

import "github.com/c4ei/YunSeokYeol/domain/consensus/model/externalapi"

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, multiset Multiset)
	IsStaged(stagingArea *StagingArea) bool
	Get(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (Multiset, error)
	Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
}
