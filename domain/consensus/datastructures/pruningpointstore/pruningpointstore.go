package pruningpointstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// PruningPointStore represents a store for the current pruning point
type PruningPointStore struct {
}

// New instantiates a new PruningPointStore
func New() *PruningPointStore {
	return &PruningPointStore{}
}

// Update updates the pruning point to be the given blockHash
func (pps *PruningPointStore) Update(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}

// Get gets the current pruning point
func (pps *PruningPointStore) Get(dbContext model.DBContextProxy) *model.DomainHash {
	return nil
}
