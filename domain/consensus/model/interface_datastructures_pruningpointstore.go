package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningPointStore represents a store for the current pruning point
type PruningPointStore interface {
	Update(dbTx TxContextProxy, blockHash *daghash.Hash)
	Get(dbContext ContextProxy) *daghash.Hash
}
