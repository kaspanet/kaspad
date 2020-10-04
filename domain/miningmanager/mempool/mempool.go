package mempool

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate"
	"github.com/kaspanet/kaspad/util"
)

// Mempool maintains a set of known transactions that
// have no yet been added to any block
type Mempool struct {
	kaspadState *kaspadstate.KaspadState
}

// New create a new Mempool
func New(kaspadState *kaspadstate.KaspadState) *Mempool {
	return &Mempool{
		kaspadState: kaspadState,
	}
}

// HandleNewBlock handles a new block that was just added to the DAG
func (mp *Mempool) HandleNewBlock(block *appmessage.MsgBlock) {

}

// Transactions returns all the transactions in the mempool
func (mp *Mempool) Transactions() []*util.Tx {
	return nil
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the mempool
func (mp *Mempool) ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error {
	return nil
}
