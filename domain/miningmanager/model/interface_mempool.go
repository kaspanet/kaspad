package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlock(block *externalapi.DomainBlock)
	Transactions() []*externalapi.DomainTransaction
	ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction) error
}
