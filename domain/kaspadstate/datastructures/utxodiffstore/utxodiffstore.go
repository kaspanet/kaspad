package utxodiffstore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type UTXODiffStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *UTXODiff)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *UTXODiff
}
