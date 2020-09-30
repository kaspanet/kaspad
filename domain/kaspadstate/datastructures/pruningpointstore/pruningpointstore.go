package pruningpointstore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningPointStore ...
type PruningPointStore struct {
}

// New ...
func New() *PruningPointStore {
	return &PruningPointStore{}
}

// Set ...
func (pps *PruningPointStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

// Get ...
func (pps *PruningPointStore) Get(dbContext dbaccess.Context) *daghash.Hash {
	return nil
}
