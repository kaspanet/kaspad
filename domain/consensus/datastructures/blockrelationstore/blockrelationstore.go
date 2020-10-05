package blockrelationstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore struct {
}

// New instantiates a new BlockRelationStore
func New() *BlockRelationStore {
	return &BlockRelationStore{}
}

// Set ...
func (brs *BlockRelationStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockRelationData *model.BlockRelations) {

}

// Get ...
func (brs *BlockRelationStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockRelations {
	return nil
}
