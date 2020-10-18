package model

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, utxoDiff *UTXODiff, utxoDiffChild *DomainHash) error
	UTXODiff(dbContext DBContextProxy, blockHash *DomainHash) (*UTXODiff, error)
	UTXODiffChild(dbContext DBContextProxy, blockHash *DomainHash) (*DomainHash, error)
	Delete(dbTx DBTxProxy, blockHash *DomainHash) error
}
