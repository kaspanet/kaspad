package model

// PruningPointStore represents a store for the current pruning point
type PruningPointStore interface {
	Update(dbTx TxContextProxy, blockHash *DomainHash)
	Get(dbContext ContextProxy) *DomainHash
}
