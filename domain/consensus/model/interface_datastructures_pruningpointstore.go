package model

// PruningPointStore represents a store for the current pruning point
type PruningPointStore interface {
	Update(dbTx DBTxProxy, blockHash *DomainHash)
	Get(dbContext DBContextProxy) *DomainHash
}
