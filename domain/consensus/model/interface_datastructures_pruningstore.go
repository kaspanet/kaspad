package model

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Update(dbTx DBTxProxy, pruningPointBlockHash *DomainHash, pruningPointUTXOSet ReadOnlyUTXOSet)
	PruningPoint(dbContext DBContextProxy) *DomainHash
	PruningPointSerializedUTXOSet(dbContext DBContextProxy) []byte
}
