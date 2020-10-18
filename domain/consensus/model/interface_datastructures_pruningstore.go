package model

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Update(dbTx DBTxProxy, pruningPointBlockHash *DomainHash, pruningPointUTXOSet ReadOnlyUTXOSet) error
	PruningPoint(dbContext DBContextProxy) (*DomainHash, error)
	PruningPointSerializedUTXOSet(dbContext DBContextProxy) ([]byte, error)
}
