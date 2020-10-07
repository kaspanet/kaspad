package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *BlockRelations)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *BlockRelations
}
