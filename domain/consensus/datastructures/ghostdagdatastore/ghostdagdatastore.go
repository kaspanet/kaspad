package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ghostdagDataStore represents a store of BlockGHOSTDAGData
type ghostdagDataStore struct {
}

// New instantiates a new GHOSTDAGDataStore
func New() model.GHOSTDAGDataStore {
	return &ghostdagDataStore{}
}

// Stage stages the given blockGHOSTDAGData for the given blockHash
func (gds *ghostdagDataStore) Stage(blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {
	panic("implement me")
}

func (gds *ghostdagDataStore) IsStaged() bool {
	panic("implement me")
}

func (gds *ghostdagDataStore) Discard() {
	panic("implement me")
}

func (gds *ghostdagDataStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// Get gets the blockGHOSTDAGData associated with the given blockHash
func (gds *ghostdagDataStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	return nil, nil
}
