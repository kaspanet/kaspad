package model

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, multiset Multiset)
	Get(dbContext ContextProxy, blockHash *DomainHash) Multiset
}
