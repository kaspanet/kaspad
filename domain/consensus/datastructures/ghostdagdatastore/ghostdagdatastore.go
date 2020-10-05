package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// GHOSTDAGDataStore ...
type GHOSTDAGDataStore struct {
}

// New ...
func New() *GHOSTDAGDataStore {
	return &GHOSTDAGDataStore{}
}

// Set ...
func (gds *GHOSTDAGDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {

}

// Get ...
func (gds *GHOSTDAGDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}
