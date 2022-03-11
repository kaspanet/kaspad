package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Store
	StageReachabilityData(stagingArea *StagingArea, blockHash *externalapi.DomainHash, reachabilityData ReachabilityData)
	StageReachabilityReindexRoot(stagingArea *StagingArea, reachabilityReindexRoot *externalapi.DomainHash)
	IsStaged(stagingArea *StagingArea) bool
	ReachabilityData(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (ReachabilityData, error)
	HasReachabilityData(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (bool, error)
	ReachabilityReindexRoot(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	Delete(dbContext DBWriter) error
}
