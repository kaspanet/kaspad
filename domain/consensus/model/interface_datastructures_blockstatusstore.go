package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, blockStatus BlockStatus)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) BlockStatus
}
