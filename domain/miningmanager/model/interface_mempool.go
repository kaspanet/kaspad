package model

import (
	consensusexternalapi "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlockTransactions(txs []*consensusexternalapi.DomainTransaction)
	Transactions() []*consensusexternalapi.DomainTransaction
	ValidateAndInsertTransaction(transaction *consensusexternalapi.DomainTransaction, allowOrphan bool) error
	RemoveTransactions(txs []*consensusexternalapi.DomainTransaction)
}
