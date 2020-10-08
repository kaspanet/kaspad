package model

import (
	"github.com/kaspanet/go-secp256k1"
)

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, multiset *secp256k1.MultiSet)
	Get(dbContext ContextProxy, blockHash *DomainHash) *secp256k1.MultiSet
}
