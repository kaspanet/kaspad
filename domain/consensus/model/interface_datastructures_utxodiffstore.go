package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *UTXODiff)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *UTXODiff
}
