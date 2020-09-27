package mempoolimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate"
	"github.com/kaspanet/kaspad/util"
)

type Mempool struct {
	kaspadState *kaspadstate.KaspadState
}

func New(kaspadState *kaspadstate.KaspadState) *Mempool {
	return &Mempool{
		kaspadState: kaspadState,
	}
}

func (mp *Mempool) HandleNewBlock(block *appmessage.MsgBlock) {

}

func (mp *Mempool) Transactions() []*util.Tx {
	return nil
}

func (mp *Mempool) ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error {
	return nil
}
