package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// MempoolTransaction represents a transaction inside the main TransactionPool
type MempoolTransaction struct {
	transaction              *externalapi.DomainTransaction
	parentTransactionsInPool IDToTransactionMap
	isHighPriority           bool
	addedAtDAAScore          uint64
}

// NewMempoolTransaction constructs a new MempoolTransaction
func NewMempoolTransaction(
	transaction *externalapi.DomainTransaction,
	parentTransactionsInPool IDToTransactionMap,
	isHighPriority bool,
	addedAtDAAScore uint64,
) *MempoolTransaction {
	return &MempoolTransaction{
		transaction:              transaction,
		parentTransactionsInPool: parentTransactionsInPool,
		isHighPriority:           isHighPriority,
		addedAtDAAScore:          addedAtDAAScore,
	}
}

// TransactionID returns the ID of this MempoolTransaction
func (mt *MempoolTransaction) TransactionID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(mt.transaction)
}

// Transaction return the DomainTransaction associated with this MempoolTransaction:
func (mt *MempoolTransaction) Transaction() *externalapi.DomainTransaction {
	return mt.transaction
}

// ParentTransactionsInPool a list of parent transactions that exist in the mempool, indexed by outpoint
func (mt *MempoolTransaction) ParentTransactionsInPool() IDToTransactionMap {
	return mt.parentTransactionsInPool
}

// RemoveParentTransactionInPool deletes a transaction from the parentTransactionsInPool set
func (mt *MempoolTransaction) RemoveParentTransactionInPool(transactionID *externalapi.DomainTransactionID) {
	delete(mt.parentTransactionsInPool, *transactionID)
}

// IsHighPriority returns whether this MempoolTransaction is a high-priority one
func (mt *MempoolTransaction) IsHighPriority() bool {
	return mt.isHighPriority
}

// AddedAtDAAScore returns the virtual DAA score at which this MempoolTransaction was added to the mempool
func (mt *MempoolTransaction) AddedAtDAAScore() uint64 {
	return mt.addedAtDAAScore
}
