package model

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, multiset Multiset)
	Get(dbContext DBContextProxy, blockHash *DomainHash) Multiset
}
