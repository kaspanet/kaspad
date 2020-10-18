package model

import (
	consensusmodel "github.com/kaspanet/kaspad/domain/consensus/model"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlock(block *consensusmodel.DomainBlock)
	Transactions() []*consensusmodel.DomainTransaction
	ValidateAndInsertTransaction(transaction *consensusmodel.DomainTransaction) error
}
