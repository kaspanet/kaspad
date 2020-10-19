package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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
func (mp *mempool) HandleNewBlock(block *externalapi.DomainBlock) {

}

// Transactions returns all the transactions in the mempool
func (mp *mempool) Transactions() []*externalapi.DomainTransaction {
	return nil
}

// ValidateAndInsertTransaction validates the given transaction, and
// adds it to the mempool
func (mp *mempool) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction) error {
	return nil
}
