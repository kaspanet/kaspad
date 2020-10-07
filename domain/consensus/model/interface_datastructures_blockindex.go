package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockIndex represents a store of known block hashes
type BlockIndex interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash)
	Exists(dbContext ContextProxy, blockHash *daghash.Hash) bool
}
