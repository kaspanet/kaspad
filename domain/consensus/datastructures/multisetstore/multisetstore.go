package multisetstore

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// MultisetStore represents a store of Multisets
type MultisetStore struct {
}

// New instantiates a new MultisetStore
func New() *MultisetStore {
	return &MultisetStore{}
}

// Insert inserts the given multiset for the given blockHash
func (ms *MultisetStore) Insert(dbTx model.TxContextProxy, blockHash *daghash.Hash, multiset *secp256k1.MultiSet) {

}

// Get gets the multiset associated with the given blockHash
func (ms *MultisetStore) Get(dbContext model.ContextProxy, blockHash *daghash.Hash) *secp256k1.MultiSet {
	return nil
}
