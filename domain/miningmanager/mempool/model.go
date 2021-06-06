package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

type idToTransaction map[externalapi.DomainTransactionID]*mempoolTransaction

type mempoolTransaction struct {
	transaction    *externalapi.DomainTransaction
	parentsInPool  idToTransaction
	isHighPriority bool
	addAtDAAScore  uint64
}

func (mt *mempoolTransaction) transactionID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(mt.transaction)
}

type orphanTransaction struct {
	transaction     *externalapi.DomainTransaction
	isHighPriority  bool
	addedAtDAAScore uint64
}

func (ot *orphanTransaction) transactionID() *externalapi.DomainTransactionID {
	return consensushashing.TransactionID(ot.transaction)
}
