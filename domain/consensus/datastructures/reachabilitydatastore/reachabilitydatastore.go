package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore struct {
}

// New instantiates a new ReachabilityDataStore
func New() *ReachabilityDataStore {
	return &ReachabilityDataStore{}
}

// Insert inserts the given reachabilityData for the given blockHash
func (rds *ReachabilityDataStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, reachabilityData *model.ReachabilityData) {

}

// Get gets the reachabilityData associated with the given blockHash
func (rds *ReachabilityDataStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.ReachabilityData {
	return nil
}
