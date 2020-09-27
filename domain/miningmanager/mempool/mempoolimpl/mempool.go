package mempoolimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/state"
	"github.com/kaspanet/kaspad/util"
)

type Mempool struct {
	state *state.State
}

func New(state *state.State) *Mempool {
	return &Mempool{
		state: state,
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
