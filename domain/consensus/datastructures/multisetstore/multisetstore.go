package multisetstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// MultisetStore represents a store of Multisets
type MultisetStore struct {
}

// New instantiates a new MultisetStore
func New() *MultisetStore {
	return &MultisetStore{}
}

// Insert inserts the given multiset for the given blockHash
func (ms *MultisetStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, multiset model.Multiset) {

}

// Get gets the multiset associated with the given blockHash
func (ms *MultisetStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) model.Multiset {
	return nil
}

// Delete deletes the multiset associated with the given blockHash
func (ms *MultisetStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
