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

// Update updates the pruning state
func (pps *PruningPointStore) Update(dbTx model.DBTxProxy, pruningPointBlockHash *model.DomainHash, pruningPointUTXOSet model.ReadOnlyUTXOSet) {

}

// PruningPoint gets the current pruning point
func (pps *PruningPointStore) PruningPoint(dbContext model.DBContextProxy) *model.DomainHash {
	return nil
}

// PruningPointSerializedUTXOSet returns the serialized UTXO set of the current pruning point
func (pps *PruningPointStore) PruningPointSerializedUTXOSet(dbContext model.DBContextProxy) []byte {
	return nil
}
