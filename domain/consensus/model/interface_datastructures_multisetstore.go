package model

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, multiset Multiset) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (Multiset, error)
	Delete(dbTx DBTxProxy, blockHash *DomainHash) error
}
