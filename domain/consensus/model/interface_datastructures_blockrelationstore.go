package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, blockRelationData *BlockRelations)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) *BlockRelations
}
