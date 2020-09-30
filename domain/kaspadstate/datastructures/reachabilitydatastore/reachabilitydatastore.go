package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type ReachabilityDataStore struct {
}

func New() *ReachabilityDataStore {
	return &ReachabilityDataStore{}
}

func (rds *ReachabilityDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, reachabilityData *model.ReachabilityData) {

}

func (rds *ReachabilityDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.ReachabilityData {
	return nil
}
