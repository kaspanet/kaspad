package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type UTXODiffStore struct {
}

func New() *UTXODiffStore {
	return &UTXODiffStore{}
}

func (uds *UTXODiffStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *model.UTXODiff) {

}

func (uds *UTXODiffStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.UTXODiff {
	return nil
}
