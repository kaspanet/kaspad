package mempoolimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate"
	"github.com/kaspanet/kaspad/util"
)

// Mempool ...
type Mempool struct {
	kaspadState *kaspadstate.KaspadState
}

// New ...
func New(kaspadState *kaspadstate.KaspadState) *Mempool {
	return &Mempool{
		kaspadState: kaspadState,
	}
}

// HandleNewBlock ...
func (mp *Mempool) HandleNewBlock(block *appmessage.MsgBlock) {

}

// Transactions ...
func (mp *Mempool) Transactions() []*util.Tx {
	return nil
}

// ValidateAndInsertTransaction ...
func (mp *Mempool) ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error {
	return nil
}
