package mempool

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type previousOutpointToOrphans map[externalapi.DomainOutpoint]idToTransaction

type orphansPool struct {
	mempool                  *mempool
	allOrphans               idToTransaction
	orphanByPreviousOutpoint previousOutpointToOrphans
	previousExpireScan       uint64
}

func newOrphansPool(mp *mempool) *orphansPool {
	return &orphansPool{
		mempool:                  mp,
		allOrphans:               idToTransaction{},
		orphanByPreviousOutpoint: previousOutpointToOrphans{},
		previousExpireScan:       0,
	}
}

func (op *orphansPool) maybeAddOrphan(transaction *externalapi.DomainTransaction,
	missingParents []*externalapi.DomainTransactionID, isHighPriority bool) error {

	panic("orphansPool.maybeAddOrphan not implemented") // TODO (Mike)
}

func (op *orphansPool) processOrphansAfterAcceptedTransaction(acceptedTransaction *mempoolTransaction) (
	acceptedOrphans []*mempoolTransaction, err error) {

	panic("orphansPool.processOrphansAfterAcceptedTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) unorphanTransaction(orphanTransactionID *externalapi.DomainTransactionID) (mempoolTransaction, error) {
	panic("orphansPool.unorphanTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) removeOrphan(orphanTransactionID *externalapi.DomainTransactionID) error {
	panic("orphansPool.removeOrphan not implemented") // TODO (Mike)
}

func (op *orphansPool) expireOrphanTransactions() error {
	panic("orphansPool.expireOrphanTransactions not implemented") // TODO (Mike)
}
