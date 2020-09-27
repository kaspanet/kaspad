package blockrelationstore

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockRelationStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *BlockRelations)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *BlockRelations
}
