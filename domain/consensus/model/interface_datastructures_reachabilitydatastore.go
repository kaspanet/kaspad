package model

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, reachabilityData *ReachabilityData)
	Get(dbContext ContextProxy, blockHash *DomainHash) *ReachabilityData
}
