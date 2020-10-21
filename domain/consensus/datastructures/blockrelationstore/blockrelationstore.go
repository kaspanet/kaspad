package blockrelationstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blockRelationStore represents a store of BlockRelations
type blockRelationStore struct {
}

// New instantiates a new BlockRelationStore
func New() model.BlockRelationStore {
	return &blockRelationStore{}
}

// Stage stages the given blockRelationData for the given blockHash
func (brs *blockRelationStore) Stage(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) error {
	panic("implement me")
}

func (brs *blockRelationStore) Discard() {
	panic("implement me")
}

func (brs *blockRelationStore) Commit(dbTx model.DBTxProxy) error {
	panic("implement me")
}

// Get gets the blockRelationData associated with the given blockHash
func (brs *blockRelationStore) Get(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*model.BlockRelations, error) {
	return nil, nil
}
