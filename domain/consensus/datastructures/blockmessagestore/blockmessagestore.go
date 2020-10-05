package blockmessagestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore struct {
}

// New instantiates a new BlockMessageStore
func New() *BlockMessageStore {
	return &BlockMessageStore{}
}

// Set ...
func (bms *BlockMessageStore) Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock) {

}

// Get ...
func (bms *BlockMessageStore) Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *appmessage.MsgBlock {
	return nil
}
