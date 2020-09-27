package blockstatusstoreimpl

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockStatusStore struct {
}

func New() *BlockStatusStore {
	return &BlockStatusStore{}
}

func (bss *BlockStatusStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockStatus blockstatusstore.BlockStatus) {

}

func (bss *BlockStatusStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) blockstatusstore.BlockStatus {
	return 0
}
