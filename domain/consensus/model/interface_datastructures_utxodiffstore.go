package model

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, utxoDiff *UTXODiff)
	Get(dbContext ContextProxy, blockHash *DomainHash) *UTXODiff
}
