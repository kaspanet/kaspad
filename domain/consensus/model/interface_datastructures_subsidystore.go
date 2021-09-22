package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SubsidyStore represents a store of block subsidies
type SubsidyStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, subsidy uint64)
	Get(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (uint64, error)
	Has(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (bool, error)
	Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
}
