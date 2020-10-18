package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
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
func (mp *Mempool) HandleNewBlock(block *consensusmodel.DomainBlock) {

}

// Transactions returns all the transactions in the mempool
func (mp *Mempool) Transactions() []*consensusmodel.DomainTransaction {
	return nil
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the mempool
func (mp *Mempool) ValidateAndInsertTransaction(transaction *consensusmodel.DomainTransaction) error {
	return nil
}
