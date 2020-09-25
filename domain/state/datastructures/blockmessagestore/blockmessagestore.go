package blockmessagestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type BlockMessageStore interface {
	Set(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, msgBlock *appmessage.MsgBlock)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *appmessage.MsgBlock
}
