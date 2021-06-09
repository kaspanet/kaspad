package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// OrphanTransaction represents a transaction in the OrphanPool
type OrphanTransaction struct {
	transaction     *externalapi.DomainTransaction
	isHighPriority  bool
	addedAtDAAScore uint64
}

func NewOrphanTransaction(
	transaction *externalapi.DomainTransaction,
	isHighPriority bool,
	addedAtDAAScore uint64,
) *OrphanTransaction {
	return &OrphanTransaction{
		transaction:     transaction,
		isHighPriority:  isHighPriority,
		addedAtDAAScore: addedAtDAAScore,
	}
}

// TransactionID returns the ID of this OrphanTransaction
func (ot *OrphanTransaction) TransactionID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(ot.transaction)
}

func (ot *OrphanTransaction) Transaction() *externalapi.DomainTransaction {
	return ot.transaction
}
func (ot *OrphanTransaction) IsHighPriority() bool {
	return ot.isHighPriority
}

func (ot *OrphanTransaction) AddedAtDAAScore() uint64 {
	return ot.addedAtDAAScore
}
