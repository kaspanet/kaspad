package mempool

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

type Mempool interface {
	HandleNewBlock(block *appmessage.MsgBlock)
	Transactions() []*util.Tx
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}
