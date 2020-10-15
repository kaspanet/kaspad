package multisetstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// multisetStore represents a store of Multisets
type multisetStore struct {
}

// New instantiates a new multisetStore
func New() model.MultisetStore {
	return &multisetStore{}
}

// Insert inserts the given multiset for the given blockHash
func (ms *multisetStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, multiset model.Multiset) {

}

// Get gets the multiset associated with the given blockHash
func (ms *multisetStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) model.Multiset {
	return nil
}

// Delete deletes the multiset associated with the given blockHash
func (ms *multisetStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
