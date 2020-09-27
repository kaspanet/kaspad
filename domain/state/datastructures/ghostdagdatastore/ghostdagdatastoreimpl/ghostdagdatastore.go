package ghostdagdatastoreimpl

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGDataStore struct {
}

func New() *GHOSTDAGDataStore {
	return &GHOSTDAGDataStore{}
}

func (gds *GHOSTDAGDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *ghostdagdatastore.BlockGHOSTDAGData) {

}

func (gds *GHOSTDAGDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *ghostdagdatastore.BlockGHOSTDAGData {
	return nil
}
