package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, msgBlock *DomainBlock)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) *DomainBlock
}
