package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// OrphanTransaction represents a transaction in the OrphanPool
type OrphanTransaction struct {
	Transaction     *externalapi.DomainTransaction
	IsHighPriority  bool
	AddedAtDAAScore uint64
}

// TransactionID returns the ID of this OrphanTransaction
func (ot *OrphanTransaction) TransactionID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(ot.Transaction)
}
