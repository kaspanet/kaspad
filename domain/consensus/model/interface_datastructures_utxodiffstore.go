package model

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, utxoDiff *UTXODiff)
	Get(dbContext DBContextProxy, blockHash *DomainHash) *UTXODiff
}
