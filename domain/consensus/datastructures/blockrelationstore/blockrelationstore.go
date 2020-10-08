package blockrelationstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore struct {
}

// New instantiates a new BlockRelationStore
func New() *BlockRelationStore {
	return &BlockRelationStore{}
}

// Insert inserts the given blockRelationData for the given blockHash
func (brs *BlockRelationStore) Insert(dbTx model.TxContextProxy, blockHash *model.DomainHash, blockRelationData *model.BlockRelations) {

}

// Get gets the blockRelationData associated with the given blockHash
func (brs *BlockRelationStore) Get(dbContext model.ContextProxy, blockHash *model.DomainHash) *model.BlockRelations {
	return nil
}
