package acceptancedatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// AcceptanceDataStore ...
type AcceptanceDataStore struct {
}

// New ...
func New() *AcceptanceDataStore {
	return &AcceptanceDataStore{}
}

// Set ...
func (ads *AcceptanceDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *model.BlockAcceptanceData) {

}

// Get ...
func (ads *AcceptanceDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockAcceptanceData {
	return nil
}
