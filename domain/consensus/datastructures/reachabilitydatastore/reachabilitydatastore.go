package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
}

// New instantiates a new ReachabilityDataStore
func New() model.ReachabilityDataStore {
	return &reachabilityDataStore{}
}

// Stage stages the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) Stage(blockHash *externalapi.DomainHash, reachabilityData *model.ReachabilityData) {
	panic("implement me")
}

func (rds *reachabilityDataStore) IsStaged() bool {
	panic("implement me")
}

func (rds *reachabilityDataStore) Discard() {
	panic("implement me")
}

func (rds *reachabilityDataStore) Commit(dbTx model.DBTxProxy) error {
	panic("implement me")
}

// Get gets the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) Get(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {
	return nil, nil
}
