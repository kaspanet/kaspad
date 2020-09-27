package utxodiffstoreimpl

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type UTXODiffStore struct {
}

func New() *UTXODiffStore {
	return &UTXODiffStore{}
}

func (uds *UTXODiffStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, utxoDiff *utxodiffstore.UTXODiff) {

}

func (uds *UTXODiffStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *utxodiffstore.UTXODiff {
	return nil
}
