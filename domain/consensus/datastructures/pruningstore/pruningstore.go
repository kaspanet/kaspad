package pruningstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// pruningStore represents a store for the current pruning state
type pruningStore struct {
}

// New instantiates a new pruningStore
func New() model.PruningStore {
	return &pruningStore{}
}

// Update updates the pruning state
func (pps *pruningStore) Update(dbTx model.DBTxProxy, pruningPointBlockHash *model.DomainHash, pruningPointUTXOSet model.ReadOnlyUTXOSet) error {
	return nil
}

// PruningPoint gets the current pruning point
func (pps *pruningStore) PruningPoint(dbContext model.DBContextProxy) (*model.DomainHash, error) {
	return nil, nil
}

// PruningPointSerializedUTXOSet returns the serialized UTXO set of the current pruning point
func (pps *pruningStore) PruningPointSerializedUTXOSet(dbContext model.DBContextProxy) ([]byte, error) {
	return nil, nil
}
