package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
)

type idToOrphan map[externalapi.DomainTransactionID]*model.OrphanTransaction
type previousOutpointToOrphans map[externalapi.DomainOutpoint]idToOrphan

type orphansPool struct {
	mempool                  *mempool
	allOrphans               idToOrphan
	orphanByPreviousOutpoint previousOutpointToOrphans
	lastExpireScan           uint64
}

func newOrphansPool(mp *mempool) *orphansPool {
	return &orphansPool{
		mempool:                  mp,
		allOrphans:               idToOrphan{},
		orphanByPreviousOutpoint: previousOutpointToOrphans{},
		lastExpireScan:           0,
	}
}

func (op *orphansPool) maybeAddOrphan(transaction *externalapi.DomainTransaction,
	missingParents []*externalapi.DomainTransactionID, isHighPriority bool) error {

	panic("orphansPool.maybeAddOrphan not implemented") // TODO (Mike)
}

func (op *orphansPool) processOrphansAfterAcceptedTransaction(acceptedTransaction *model.MempoolTransaction) (
	acceptedOrphans []*model.MempoolTransaction, err error) {

	panic("orphansPool.processOrphansAfterAcceptedTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) unorphanTransaction(orphanTransactionID *externalapi.DomainTransactionID) (model.MempoolTransaction, error) {
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
		if orphanTransaction.IsHighPriority {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then transactionExpireIntervalDAAScore
		if virtualDAAScore-orphanTransaction.AddedAtDAAScore > op.mempool.config.orphanExpireIntervalDAAScore {
			err = op.removeOrphan(orphanTransaction.TransactionID())
			if err != nil {
				return err
			}
		}
	}

	op.lastExpireScan = virtualDAAScore
	return nil
}
