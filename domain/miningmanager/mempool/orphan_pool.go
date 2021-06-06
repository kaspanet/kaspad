package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

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
	missingParents []*externalapi.DomainTransactionID, neverExpires bool) error {

	panic("orphansPool.maybeAddOrphan not implemented") // TODO (Mike)
}

func (op *orphansPool) processOrphansAfterAcceptedTransaction(acceptedTransaction *mempoolTransaction) (
	acceptedOrphans []*mempoolTransaction, err error) {

	panic("orphansPool.processOrphansAfterAcceptedTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) unorphanTransaction(orphanTransactionID *externalapi.DomainTransactionID) (mempoolTransaction, error) {
	panic("orphansPool.unorphanTransaction not implemented") // TODO (Mike)
}

func (op *orphansPool) removeOrphan(orphanTransactionID *externalapi.DomainTransactionID, removeRedeemers bool) {
	var orphanTransaction *orphanTransaction
	var ok bool
	if orphanTransaction, ok = op.allOrphans[*orphanTransactionID]; !ok {
		return
	}

	// Remove orphan from allOrphans
	delete(op.allOrphans, *orphanTransactionID)

	// Remove orphan from all relevant entries in orphanByPreviousOutpoint
	for _, input := range orphanTransaction.transaction.Inputs {
		if orphans, ok := op.orphanByPreviousOutpoint[input.PreviousOutpoint]; ok {
			delete(orphans, orphanTransactionID)
			if len(orphans) == 0 {
				delete(op.orphanByPreviousOutpoint, input.PreviousOutpoint)
			}
		}
	}

	// Recursively remove redeemers if requested.
	// Since the size of the orphan pool is very limited - the recursion depth is properly bound.
	if removeRedeemers {
		outpoint := externalapi.DomainOutpoint{TransactionID: *orphanTransactionID}
		for i := range orphanTransaction.transaction.Outputs {
			outpoint.Index = uint32(i)
			for _, orphanRedeemer := range op.orphanByPreviousOutpoint[outpoint] {
				op.removeOrphan(orphanRedeemer.transactionID(), true)
			}
		}
	}
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
		if orphanTransaction.neverExpires {
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
