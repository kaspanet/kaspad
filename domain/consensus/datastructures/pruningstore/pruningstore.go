package pruningstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// pruningStore represents a store for the current pruning state
type pruningStore struct {
}

// New instantiates a new PruningStore
func New() model.PruningStore {
	return &pruningStore{}
}

// Stage stages the pruning state
func (pps *pruningStore) Stage(pruningPointBlockHash *externalapi.DomainHash, pruningPointUTXOSet model.ReadOnlyUTXOSet) {
	panic("implement me")
}

func (pps *pruningStore) IsStaged() bool {
	panic("implement me")
}

func (pps *pruningStore) Discard() {
	panic("implement me")
}

func (pps *pruningStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// PruningPoint gets the current pruning point
func (pps *pruningStore) PruningPoint(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	return nil, nil
}

// PruningPointSerializedUTXOSet returns the serialized UTXO set of the current pruning point
func (pps *pruningStore) PruningPointSerializedUTXOSet(dbContext model.DBReader) ([]byte, error) {
	return nil, nil
}
