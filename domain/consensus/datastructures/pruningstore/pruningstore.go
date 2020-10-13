package pruningstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// PruningStore represents a store for the current pruning state
type PruningStore struct {
}

// New instantiates a new PruningStore
func New() *PruningStore {
	return &PruningStore{}
}

// Update updates the pruning state
func (pps *PruningStore) Update(dbTx model.DBTxProxy, pruningPointBlockHash *model.DomainHash, pruningPointUTXOSet model.ReadOnlyUTXOSet) {

}

// PruningPoint gets the current pruning point
func (pps *PruningStore) PruningPoint(dbContext model.DBContextProxy) *model.DomainHash {
	return nil
}

// PruningPointSerializedUTXOSet returns the serialized UTXO set of the current pruning point
func (pps *PruningStore) PruningPointSerializedUTXOSet(dbContext model.DBContextProxy) []byte {
	return nil
}
