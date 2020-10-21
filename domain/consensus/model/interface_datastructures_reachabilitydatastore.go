package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	StageReachabilityData(blockHash *externalapi.DomainHash, reachabilityData *ReachabilityData)
	StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash)
	IsAnythingStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	ReachabilityData(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*ReachabilityData, error)
	ReachabilityReindexRoot(dbContext DBContextProxy) (*externalapi.DomainHash, error)
}
