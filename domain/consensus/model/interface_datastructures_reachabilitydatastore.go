package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Store
	StageReachabilityData(blockHash *externalapi.DomainHash, reachabilityData *ReachabilityData)
	StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash)
	IsAnythingStaged() bool
	ReachabilityData(dbContext DBReader, blockHash *externalapi.DomainHash) (*ReachabilityData, error)
	ReachabilityReindexRoot(dbContext DBReader) (*externalapi.DomainHash, error)
}
