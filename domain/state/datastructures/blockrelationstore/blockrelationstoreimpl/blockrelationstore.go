package blockrelationstoreimpl

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockRelationStore struct {
}

func New() *BlockRelationStore {
	return &BlockRelationStore{}
}

func (brs *BlockRelationStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *blockrelationstore.BlockRelations) {

}

func (brs *BlockRelationStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *blockrelationstore.BlockRelations {
	return nil
}
