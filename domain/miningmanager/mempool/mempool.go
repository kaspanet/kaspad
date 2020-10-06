package mempool

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/util"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool struct {
	consensus *consensus.Consensus
}

// New create a new Mempool
func New(consensus *consensus.Consensus) *Mempool {
	return &Mempool{
		consensus: consensus,
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
