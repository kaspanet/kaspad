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

// Set ...
func (rds *ReachabilityDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData) {

}

// Get ...
func (rds *ReachabilityDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.ReachabilityData {
	return nil
}
