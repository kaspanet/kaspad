package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/estimatedsize"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/pkg/errors"
)

type idToOrphan map[externalapi.DomainTransactionID]*model.OrphanTransaction
type previousOutpointToOrphans map[externalapi.DomainOutpoint]idToOrphan

type orphansPool struct {
	mempool                   *mempool
	allOrphans                idToOrphan
	orphansByPreviousOutpoint previousOutpointToOrphans
	lastExpireScan            uint64
}

func newOrphansPool(mp *mempool) *orphansPool {
	return &orphansPool{
		mempool:                   mp,
		allOrphans:                idToOrphan{},
		orphansByPreviousOutpoint: previousOutpointToOrphans{},
		lastExpireScan:            0,
	}
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) maybeAddOrphan(transaction *externalapi.DomainTransaction, isHighPriority bool) error {
	serializedLength := estimatedsize.TransactionEstimatedSerializedSize(transaction)
	if serializedLength > uint64(op.mempool.config.maximumOrphanTransactionSize) {
		str := fmt.Sprintf("orphan transaction size of %d bytes is "+
			"larger than max allowed size of %d bytes",
			serializedLength, op.mempool.config.maximumOrphanTransactionSize)
		return transactionRuleError(RejectBadOrphan, str)
	}
	if op.mempool.config.maximumOrphanTransactionCount <= 0 {
		return nil
	}
	for len(op.allOrphans) >= op.mempool.config.maximumOrphanTransactionCount {
		// Don't remove redeemers in the case of a random eviction since
		// it is quite possible it might be needed again shortly.
		err := op.removeOrphan(op.randomOrphan().TransactionID(), false)
		if err != nil {
			return err
		}
	}

	return op.addOrphan(transaction, isHighPriority)
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) addOrphan(transaction *externalapi.DomainTransaction, isHighPriority bool) error {
	virtualDAAScore, err := op.mempool.virtualDAAScore()
	if err != nil {
		return err
	}
	orphanTransaction := model.NewOrphanTransaction(transaction, isHighPriority, virtualDAAScore)

	op.allOrphans[*orphanTransaction.TransactionID()] = orphanTransaction
	for _, input := range transaction.Inputs {
		if input.UTXOEntry == nil {
			if _, ok := op.orphansByPreviousOutpoint[input.PreviousOutpoint]; !ok {
				op.orphansByPreviousOutpoint[input.PreviousOutpoint] = idToOrphan{}
			}
			op.orphansByPreviousOutpoint[input.PreviousOutpoint][*orphanTransaction.TransactionID()] = orphanTransaction
		}
	}

	return nil
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) processOrphansAfterAcceptedTransaction(acceptedTransaction *externalapi.DomainTransaction) (
	acceptedOrphans []*externalapi.DomainTransaction, err error) {

	acceptedOrphans = []*externalapi.DomainTransaction{}
	queue := []*externalapi.DomainTransaction{acceptedTransaction}

	for len(queue) > 0 {
		var current *externalapi.DomainTransaction
		current, queue = queue[0], queue[1:]

		currentTransactionID := consensushashing.TransactionID(current)
		outpoint := externalapi.DomainOutpoint{TransactionID: *currentTransactionID}
		for i, output := range current.Outputs {
			outpoint.Index = uint32(i)
			orphans, ok := op.orphansByPreviousOutpoint[outpoint]
			if !ok {
				continue
			}
			for _, orphan := range orphans {
				for _, input := range orphan.Transaction().Inputs {
					if input.PreviousOutpoint.Equal(&outpoint) {
						input.UTXOEntry = utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, false,
							model.UnacceptedDAAScore)
						break
					}
				}
				if countUnfilledInputs(orphan) == 0 {
					err := op.unorphanTransaction(orphan)
					if err != nil {
						if errors.As(err, &RuleError{}) {
							log.Infof("Failed to unorphan transaction %s due to rule error: %s",
								currentTransactionID, err)
							continue
						}
						return nil, err
					}
					acceptedOrphans = append(acceptedOrphans, current)
				}
			}
		}
	}

	return acceptedOrphans, nil
}

func countUnfilledInputs(orphan *model.OrphanTransaction) int {
	unfilledInputs := 0
	for _, input := range orphan.Transaction().Inputs {
		if input.UTXOEntry == nil {
			unfilledInputs++
		}
	}
	return unfilledInputs
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) unorphanTransaction(transaction *model.OrphanTransaction) error {
	err := op.removeOrphan(transaction.TransactionID(), false)
	if err != nil {
		return err
	}

	err = op.mempool.validateTransactionInContext(transaction.Transaction())
	if err != nil {
		return err
	}

	virtualDAAScore, err := op.mempool.virtualDAAScore()
	if err != nil {
		return err
	}
	mempoolTransaction := model.NewMempoolTransaction(
		transaction.Transaction(),
		op.mempool.transactionsPool.getParentTransactionsInPool(transaction.Transaction()),
		false,
		virtualDAAScore,
	)
	err = op.mempool.transactionsPool.addMempoolTransaction(mempoolTransaction)
	if err != nil {
		return err
	}

	return nil
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) removeOrphan(orphanTransactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	orphanTransaction, ok := op.allOrphans[*orphanTransactionID]
	if !ok {
		return nil
	}

	delete(op.allOrphans, *orphanTransactionID)

	for i, input := range orphanTransaction.Transaction().Inputs {
		orphans, ok := op.orphansByPreviousOutpoint[input.PreviousOutpoint]
		if !ok {
			return errors.Errorf("Input No. %d of %s (%s) doesn't exist in orphansByPreviousOutpoint",
				i, orphanTransactionID, input.PreviousOutpoint)
		}
		delete(orphans, *orphanTransactionID)
		if len(orphans) == 0 {
			delete(op.orphansByPreviousOutpoint, input.PreviousOutpoint)
		}
	}

	if removeRedeemers {
		err := op.removeRedeemersOf(orphanTransaction)
		if err != nil {
			return err
		}
	}

	return nil
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) removeRedeemersOf(transaction model.Transaction) error {
	outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}
	for i := range transaction.Transaction().Outputs {
		outpoint.Index = uint32(i)
		for _, orphan := range op.orphansByPreviousOutpoint[outpoint] {
			// Recursive call is bound by size of orphan pool (which is very small)
			err := op.removeOrphan(orphan.TransactionID(), true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// this function MUST be called with the mempool mutex locked for writes
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
		if orphanTransaction.IsHighPriority() {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then transactionExpireIntervalDAAScore
		if virtualDAAScore-orphanTransaction.AddedAtDAAScore() > op.mempool.config.orphanExpireIntervalDAAScore {
			err = op.removeOrphan(orphanTransaction.TransactionID(), true)
			if err != nil {
				return err
			}
		}
	}

	op.lastExpireScan = virtualDAAScore
	return nil
}

// this function MUST be called with the mempool mutex locked for writes
func (op *orphansPool) updateOrphansAfterTransactionRemoved(
	removedTransaction *model.MempoolTransaction, removeRedeemers bool) error {

	if removeRedeemers {
		return op.removeRedeemersOf(removedTransaction)
	}

	outpoint := externalapi.DomainOutpoint{TransactionID: *removedTransaction.TransactionID()}
	for i := range removedTransaction.Transaction().Outputs {
		outpoint.Index = uint32(i)
		for _, orphan := range op.orphansByPreviousOutpoint[outpoint] {
			for _, input := range orphan.Transaction().Inputs {
				if input.PreviousOutpoint.TransactionID.Equal(removedTransaction.TransactionID()) {
					input.UTXOEntry = nil
				}
			}
		}
	}

	return nil
}

// this function MUST be called with the mempool mutex locked for reads
func (op *orphansPool) randomOrphan() *model.OrphanTransaction {
	for _, orphan := range op.allOrphans {
		return orphan
	}

	return nil
}
