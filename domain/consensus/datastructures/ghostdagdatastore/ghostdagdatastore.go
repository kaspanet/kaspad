package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore struct {
}

// New instantiates a new GHOSTDAGDataStore
func New() *GHOSTDAGDataStore {
	return &GHOSTDAGDataStore{}
}

// Insert inserts the given blockGHOSTDAGData for the given blockHash
func (gds *GHOSTDAGDataStore) Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {

}

// Get gets the blockGHOSTDAGData associated with the given blockHash
func (gds *GHOSTDAGDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}
