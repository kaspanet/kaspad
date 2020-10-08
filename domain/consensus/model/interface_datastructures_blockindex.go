package model

// BlockIndex represents a store of known block hashes
type BlockIndex interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash)
	Exists(dbContext ContextProxy, blockHash *DomainHash) bool
}
