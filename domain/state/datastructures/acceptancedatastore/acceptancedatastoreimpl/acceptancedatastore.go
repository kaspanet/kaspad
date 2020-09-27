package acceptancedatastoreimpl

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type AcceptanceDataStore struct {
}

func New() *AcceptanceDataStore {
	return &AcceptanceDataStore{}
}

func (ads *AcceptanceDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *acceptancedatastore.AcceptanceData) {

}

func (ads *AcceptanceDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *acceptancedatastore.AcceptanceData {
	return nil
}
