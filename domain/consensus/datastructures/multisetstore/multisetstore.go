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
func (ms *MultisetStore) Insert(dbTx model.TxContextProxy, blockHash *model.DomainHash, multiset model.Multiset) {

}

// Get gets the multiset associated with the given blockHash
func (ms *MultisetStore) Get(dbContext model.ContextProxy, blockHash *model.DomainHash) model.Multiset {
	return nil
}
