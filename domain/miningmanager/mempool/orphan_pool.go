package mempool

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type previousOutpointToOrphans map[externalapi.DomainOutpoint]idToTransaction

type orphansPool struct {
	mempool                  *mempool
	allOrphans               idToTransaction
	orphanByPreviousOutpoint previousOutpointToOrphans
	lastExpireScan           uint64
}

func newOrphansPool(mp *mempool) *orphansPool {
	return &orphansPool{
		mempool:                  mp,
		allOrphans:               idToTransaction{},
		orphanByPreviousOutpoint: previousOutpointToOrphans{},
		lastExpireScan:           0,
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
	virtualDAAScore, err := op.mempool.virtualDAAScore()
	if err != nil {
		return err
	}

	if virtualDAAScore-op.lastExpireScan < op.mempool.config.orphanExpireScanIntervalDAAScore {
		return nil
	}

	for _, orphanTransaction := range op.allOrphans {
		// Never expire high priority transactions
		if orphanTransaction.isHighPriority {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then transactionExpireIntervalDAAScore
		if virtualDAAScore-orphanTransaction.addAtDAAScore > op.mempool.config.orphanExpireIntervalDAAScore {
			err = op.removeOrphan(orphanTransaction.transactionID())
			if err != nil {
				return err
			}
		}
	}

	op.lastExpireScan = virtualDAAScore
	return nil
}
