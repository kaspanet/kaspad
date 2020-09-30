package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type AcceptanceDataStore struct {
}

func New() *AcceptanceDataStore {
	return &AcceptanceDataStore{}
}

func (ads *AcceptanceDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *model.AcceptanceData) {

}

func (ads *AcceptanceDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.AcceptanceData {
	return nil
}
