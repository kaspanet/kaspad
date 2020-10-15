package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
}

// New instantiates a new reachabilityDataStore
func New() model.ReachabilityDataStore {
	return &reachabilityDataStore{}
}

// Insert inserts the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, reachabilityData *model.ReachabilityData) {

}

// Get gets the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.ReachabilityData {
	return nil
}
