package model

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, reachabilityData *ReachabilityData) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (*ReachabilityData, error)
}
