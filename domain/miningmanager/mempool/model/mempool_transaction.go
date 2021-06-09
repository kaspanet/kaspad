package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// MempoolTransaction represents a transaction inside the main TransactionPool
type MempoolTransaction struct {
	transaction              *externalapi.DomainTransaction
	parentTransactionsInPool OutpointToTransaction
	isHighPriority           bool
	addedAtDAAScore          uint64
}

func NewMempoolTransaction(
	transaction *externalapi.DomainTransaction,
	parentTransactionsInPool OutpointToTransaction,
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

func (mt *MempoolTransaction) Transaction() *externalapi.DomainTransaction {
	return mt.transaction
}

func (mt *MempoolTransaction) ParentTransactionsInPool() OutpointToTransaction {
	return mt.parentTransactionsInPool
}

func (mt *MempoolTransaction) IsHighPriority() bool {
	return mt.isHighPriority
}

func (mt *MempoolTransaction) AddedAtDAAScore() uint64 {
	return mt.addedAtDAAScore
}
