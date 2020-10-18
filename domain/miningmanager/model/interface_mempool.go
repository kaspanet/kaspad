package model

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlock(block *appmessage.MsgBlock)
	Transactions() []*util.Tx
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}
