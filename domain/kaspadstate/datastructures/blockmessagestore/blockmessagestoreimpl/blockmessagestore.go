package blockmessagestoreimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockMessageStore struct {
}

func New() *BlockMessageStore {
	return &BlockMessageStore{}
}

func (bms *BlockMessageStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock) {

}

func (bms *BlockMessageStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *appmessage.MsgBlock {
	return nil
}
