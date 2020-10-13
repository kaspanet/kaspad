package model

// FeeDataStore represents a store of fee data
type FeeDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, fee uint64)
	Get(dbContext DBContextProxy, blockHash *DomainHash) uint64
}
