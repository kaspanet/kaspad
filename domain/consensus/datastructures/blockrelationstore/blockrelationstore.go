package blockrelationstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// blockRelationStore represents a store of BlockRelations
type blockRelationStore struct {
}

// New instantiates a new BlockRelationStore
func New() model.BlockRelationStore {
	return &blockRelationStore{}
}

// Insert inserts the given blockRelationData for the given blockHash
func (brs *blockRelationStore) Update(dbTx model.DBTxProxy, blockHash *model.DomainHash, parentHashes []*model.DomainHash) error {
	return nil
}

// Get gets the blockRelationData associated with the given blockHash
func (brs *blockRelationStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) (*model.BlockRelations, error) {
	return nil, nil
}
