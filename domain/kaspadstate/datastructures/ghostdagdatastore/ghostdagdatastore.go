package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGDataStore struct {
}

func New() *GHOSTDAGDataStore {
	return &GHOSTDAGDataStore{}
}

func (gds *GHOSTDAGDataStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {

}

func (gds *GHOSTDAGDataStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}
