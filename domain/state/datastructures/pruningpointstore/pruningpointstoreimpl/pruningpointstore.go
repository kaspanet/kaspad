package pruningpointstoreimpl

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type PruningPointStore struct {
}

func New() *PruningPointStore {
	return &PruningPointStore{}
}

func (pps *PruningPointStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

func (pps *PruningPointStore) Get(dbContext dbaccess.Context) *daghash.Hash {
	return nil
}
