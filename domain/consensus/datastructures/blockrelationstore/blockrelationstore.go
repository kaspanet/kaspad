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

func (brs *blockRelationStore) StageBlockRelation(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash) {
	panic("implement me")
}

func (brs *blockRelationStore) StageTips(tipHashess []*externalapi.DomainHash) {
	panic("implement me")
}

func (brs *blockRelationStore) IsAnythingStaged() bool {
	panic("implement me")
}

func (brs *blockRelationStore) Discard() {
	panic("implement me")
}

func (brs *blockRelationStore) Commit(dbTx model.DBTxProxy) error {
	panic("implement me")
}

func (brs *blockRelationStore) BlockRelation(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*model.BlockRelations, error) {
	panic("implement me")
}

func (brs *blockRelationStore) Tips(dbContext model.DBContextProxy) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}
