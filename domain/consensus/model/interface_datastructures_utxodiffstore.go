package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, utxoDiff *UTXODiff)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) *UTXODiff
}
