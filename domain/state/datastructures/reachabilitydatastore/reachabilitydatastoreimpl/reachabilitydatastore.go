package reachabilitydatastoreimpl

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type ReachabilityDataStore struct {
}

func New() *ReachabilityDataStore {
	return &ReachabilityDataStore{}
}

func (rds *ReachabilityDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *reachabilitydatastore.ReachabilityData) {

}

func (rds *ReachabilityDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *reachabilitydatastore.ReachabilityData {
	return nil
}
