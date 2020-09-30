package blockstatusstore

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockStatusStore struct {
}

func New() *BlockStatusStore {
	return &BlockStatusStore{}
}

func (bss *BlockStatusStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus model.BlockStatus) {

}

func (bss *BlockStatusStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) model.BlockStatus {
	return 0
}
