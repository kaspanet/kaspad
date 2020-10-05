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

// Update updates the pruning point to be the given blockHash
func (pps *PruningPointStore) Update(dbTx *dbaccess.TxContext, blockHash *daghash.Hash) {

}

// Get gets the current pruning point
func (pps *PruningPointStore) Get(dbContext dbaccess.Context) *daghash.Hash {
	return nil
}
