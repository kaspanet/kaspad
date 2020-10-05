package pruningpointstore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningPointStore represents a store for the current pruning point
type PruningPointStore struct {
}

// New instantiates a new PruningPointStore
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
