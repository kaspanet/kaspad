package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Store
	StageReachabilityData(blockHash *externalapi.DomainHash, reachabilityData ReachabilityData)
	StageInterval(blockHash *externalapi.DomainHash, interval *ReachabilityInterval)
	StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash)
	IsAnythingStaged() bool
	ReachabilityData(dbContext DBReader, blockHash *externalapi.DomainHash) (ReachabilityData, error)
	Interval(dbContext DBReader, blockHash *externalapi.DomainHash) (*ReachabilityInterval, error)
	HasReachabilityData(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
	ReachabilityReindexRoot(dbContext DBReader) (*externalapi.DomainHash, error)
}
