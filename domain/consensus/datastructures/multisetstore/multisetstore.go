package multisetstore

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// MultisetStore represents a store of Multisets
type MultisetStore struct {
}

// New instantiates a new MultisetStore
func New() *MultisetStore {
	return &MultisetStore{}
}

// Set ...
func (ms *MultisetStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, multiset *secp256k1.MultiSet) {

}

// Get ...
func (ms *MultisetStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *secp256k1.MultiSet {
	return nil
}
