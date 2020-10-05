package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// UTXODiffStore ...
type UTXODiffStore struct {
}

// New ...
func New() *UTXODiffStore {
	return &UTXODiffStore{}
}

// Set ...
func (uds *UTXODiffStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *model.UTXODiff) {

}

// Get ...
func (uds *UTXODiffStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.UTXODiff {
	return nil
}
