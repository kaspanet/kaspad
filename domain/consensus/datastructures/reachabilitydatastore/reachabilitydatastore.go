package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
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
func (rds *ReachabilityDataStore) Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData) {

}

// Get gets the reachabilityData associated with the given blockHash
func (rds *ReachabilityDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.ReachabilityData {
	return nil
}
