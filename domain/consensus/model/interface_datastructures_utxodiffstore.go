package model

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, utxoDiff *UTXODiff, utxoDiffChild *DomainHash)
	UTXODiff(dbContext DBContextProxy, blockHash *DomainHash) *UTXODiff
	UTXODiffChild(dbContext DBContextProxy, blockHash *DomainHash) *DomainHash
	Delete(dbTx DBTxProxy, blockHash *DomainHash)
}
