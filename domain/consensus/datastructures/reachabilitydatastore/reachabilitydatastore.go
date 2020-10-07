package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore struct {
}

// New instantiates a new ReachabilityDataStore
func New() *ReachabilityDataStore {
	return &ReachabilityDataStore{}
}

// Insert inserts the given reachabilityData for the given blockHash
func (rds *ReachabilityDataStore) Insert(dbTx model.TxContextProxy, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData) {

}

// Get gets the reachabilityData associated with the given blockHash
func (rds *ReachabilityDataStore) Get(dbContext model.ContextProxy, blockHash *daghash.Hash) *model.ReachabilityData {
	return nil
}
