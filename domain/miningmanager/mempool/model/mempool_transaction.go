package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// MempoolTransaction represents a transaction inside the main TransactionPool
type MempoolTransaction struct {
	Transaction              *externalapi.DomainTransaction
	ParentTransactionsInPool OutpointToTransaction
	IsHighPriority           bool
	AddedAtDAAScore          uint64
}

// TransactionID returns the ID of this MempoolTransaction
func (mt *MempoolTransaction) TransactionID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(mt.Transaction)
}
