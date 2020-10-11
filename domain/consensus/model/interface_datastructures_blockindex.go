package model

// BlockIndex represents a store of known block hashes
type BlockIndex interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash)
	Exists(dbContext DBContextProxy, blockHash *DomainHash) bool
}
