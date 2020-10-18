package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/miningmanager/model"
)

// mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type mempool struct {
	consensus *consensus.Consensus
}

// New creates a new mempool
func New(consensus *consensus.Consensus) model.Mempool {
	return &mempool{
		consensus: consensus,
	}
}

// HandleNewBlock handles a new block that was just added to the DAG
func (mp *mempool) HandleNewBlock(block *consensusmodel.DomainBlock) {

}

// Transactions returns all the transactions in the mempool
func (mp *mempool) Transactions() []*consensusmodel.DomainTransaction {
	return nil
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the mempool
func (mp *mempool) ValidateAndInsertTransaction(transaction *consensusmodel.DomainTransaction) error {
	return nil
}
