package model

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, multiset *secp256k1.MultiSet)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *secp256k1.MultiSet
}
