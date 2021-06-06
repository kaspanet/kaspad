package mempool

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type idToTransaction map[externalapi.DomainTransactionID]*mempoolTransaction

type mempoolTransaction struct {
	transaction    externalapi.DomainTransaction
	parentsInPool  idToTransaction
	isHighPriority bool
	addAtDAAScore  uint64
}

type orphanTransaction struct {
	transaction     externalapi.DomainTransaction
	isHighPriority  bool
	addedAtDAAScore uint64
}
