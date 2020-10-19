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

// Insert inserts the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) Insert(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash, reachabilityData *model.ReachabilityData) error {
	return nil
}

// Get gets the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) Get(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {
	return nil, nil
}
