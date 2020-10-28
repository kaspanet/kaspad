package multisetstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// multisetStore represents a store of Multisets
type multisetStore struct {
}

// New instantiates a new MultisetStore
func New() model.MultisetStore {
	return &multisetStore{}
}

// Stage stages the given multiset for the given blockHash
func (ms *multisetStore) Stage(blockHash *externalapi.DomainHash, multiset model.Multiset) {
	panic("implement me")
}

func (ms *multisetStore) IsStaged() bool {
	panic("implement me")
}

func (ms *multisetStore) Discard() {
	panic("implement me")
}

func (ms *multisetStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// Get gets the multiset associated with the given blockHash
func (ms *multisetStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (model.Multiset, error) {
	return nil, nil
}

// Delete deletes the multiset associated with the given blockHash
func (ms *multisetStore) Delete(dbTx model.DBTransaction, blockHash *externalapi.DomainHash) error {
	return nil
}
