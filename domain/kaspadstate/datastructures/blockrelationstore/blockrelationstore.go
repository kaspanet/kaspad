package blockrelationstore

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockRelationStore struct {
}

func New() *BlockRelationStore {
	return &BlockRelationStore{}
}

func (brs *BlockRelationStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *model.BlockRelations) {

}

func (brs *BlockRelationStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockRelations {
	return nil
}
