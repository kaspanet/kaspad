package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/pkg/errors"
)

type idToOrphanMap map[externalapi.DomainTransactionID]*model.OrphanTransaction
type previousOutpointToOrphanMap map[externalapi.DomainOutpoint]*model.OrphanTransaction

type orphansPool struct {
	mempool                   *mempool
	allOrphans                idToOrphanMap
	orphansByPreviousOutpoint previousOutpointToOrphanMap
	lastExpireScan            uint64
}

func newOrphansPool(mp *mempool) *orphansPool {
	return &orphansPool{
		mempool:                   mp,
		allOrphans:                idToOrphanMap{},
		orphansByPreviousOutpoint: previousOutpointToOrphanMap{},
		lastExpireScan:            0,
	}
}

func (op *orphansPool) maybeAddOrphan(transaction *externalapi.DomainTransaction, isHighPriority bool) error {
	if op.mempool.config.MaximumOrphanTransactionCount == 0 {
		return nil
	}

	err := op.checkOrphanDuplicate(transaction)
	if err != nil {
		return err
	}

	err = op.checkOrphanMass(transaction)
	if err != nil {
		return err
	}
	err = op.checkOrphanDoubleSpend(transaction)
	if err != nil {
		return err
	}

	err = op.addOrphan(transaction, isHighPriority)
	if err != nil {
		return err
	}

	err = op.limitOrphanPoolSize()
	if err != nil {
		return err
	}

	return nil
}

func (op *orphansPool) limitOrphanPoolSize() error {
	for uint64(len(op.allOrphans)) > op.mempool.config.MaximumOrphanTransactionCount {
		orphanToRemove := op.randomNonHighPriorityOrphan()
		if orphanToRemove == nil { // this means all orphans are HighPriority
			log.Warnf(
				"Number of high-priority transactions in orphanPool (%d) is higher than maximum allowed (%d)",
				len(op.allOrphans),
				op.mempool.config.MaximumOrphanTransactionCount)
			break
		}

		// Don't remove redeemers in the case of a random eviction since the evicted transaction is
		// not invalid, therefore it's redeemers are as good as any orphan that just arrived.
		err := op.removeOrphan(orphanToRemove.TransactionID(), false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (op *orphansPool) checkOrphanMass(transaction *externalapi.DomainTransaction) error {
	if transaction.Mass > op.mempool.config.MaximumOrphanTransactionMass {
		str := fmt.Sprintf("orphan transaction size of %d bytes is "+
			"larger than max allowed size of %d bytes",
			transaction.Mass, op.mempool.config.MaximumOrphanTransactionMass)
		return transactionRuleError(RejectBadOrphan, str)
	}
	return nil
}

func (op *orphansPool) checkOrphanDuplicate(transaction *externalapi.DomainTransaction) error {
	if _, ok := op.allOrphans[*consensushashing.TransactionID(transaction)]; ok {
		str := fmt.Sprintf("Orphan transacion %s is already in the orphan pool",
			consensushashing.TransactionID(transaction))
		return transactionRuleError(RejectDuplicate, str)
	}
	return nil
}

func (op *orphansPool) checkOrphanDoubleSpend(transaction *externalapi.DomainTransaction) error {
	for _, input := range transaction.Inputs {
		if doubleSpendOrphan, ok := op.orphansByPreviousOutpoint[input.PreviousOutpoint]; ok {
			str := fmt.Sprintf("Orphan transacion %s is double spending an input from already existing orphan %s",
				consensushashing.TransactionID(transaction), doubleSpendOrphan.TransactionID())
			return transactionRuleError(RejectDuplicate, str)
		}
	}

	return nil
}

func (op *orphansPool) addOrphan(transaction *externalapi.DomainTransaction, isHighPriority bool) error {
	virtualDAAScore, err := op.mempool.consensus.Consensus().GetVirtualDAAScore()
	if err != nil {
		return err
	}
	orphanTransaction := model.NewOrphanTransaction(transaction, isHighPriority, virtualDAAScore)

	op.allOrphans[*orphanTransaction.TransactionID()] = orphanTransaction
	for _, input := range transaction.Inputs {
		op.orphansByPreviousOutpoint[input.PreviousOutpoint] = orphanTransaction
	}

	return nil
}

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
			orphan, ok := op.orphansByPreviousOutpoint[outpoint]
			if !ok {
				continue
			}
			for _, input := range orphan.Transaction().Inputs {
				if input.PreviousOutpoint.Equal(&outpoint) && input.UTXOEntry == nil {
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
				acceptedOrphans = append(acceptedOrphans, orphan.Transaction())
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

func (op *orphansPool) unorphanTransaction(transaction *model.OrphanTransaction) error {
	err := op.removeOrphan(transaction.TransactionID(), false)
	if err != nil {
		return err
	}

	err = op.mempool.consensus.Consensus().ValidateTransactionAndPopulateWithConsensusData(transaction.Transaction())
	if err != nil {
		if errors.Is(err, ruleerrors.ErrImmatureSpend) {
			return transactionRuleError(RejectImmatureSpend, "one of the transaction inputs spends an immature UTXO")
		}
		if errors.As(err, &ruleerrors.RuleError{}) {
			return newRuleError(err)
		}
		return err
	}

	err = op.mempool.validateTransactionInContext(transaction.Transaction())
	if err != nil {
		return err
	}

	virtualDAAScore, err := op.mempool.consensus.Consensus().GetVirtualDAAScore()
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

func (op *orphansPool) removeOrphan(orphanTransactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	orphanTransaction, ok := op.allOrphans[*orphanTransactionID]
	if !ok {
		return nil
	}

	delete(op.allOrphans, *orphanTransactionID)

	for i, input := range orphanTransaction.Transaction().Inputs {
		if _, ok := op.orphansByPreviousOutpoint[input.PreviousOutpoint]; !ok {
			return errors.Errorf("Input No. %d of %s (%s) doesn't exist in orphansByPreviousOutpoint",
				i, orphanTransactionID, input.PreviousOutpoint)
		}
		delete(op.orphansByPreviousOutpoint, input.PreviousOutpoint)
	}

	if removeRedeemers {
		err := op.removeRedeemersOf(orphanTransaction)
		if err != nil {
			return err
		}
	}

	return nil
}

func (op *orphansPool) removeRedeemersOf(transaction model.Transaction) error {
	outpoint := externalapi.DomainOutpoint{TransactionID: *transaction.TransactionID()}
	for i := range transaction.Transaction().Outputs {
		outpoint.Index = uint32(i)
		if orphan, ok := op.orphansByPreviousOutpoint[outpoint]; ok {
			// Recursive call is bound by size of orphan pool (which is very small)
			err := op.removeOrphan(orphan.TransactionID(), true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (op *orphansPool) expireOrphanTransactions() error {
	virtualDAAScore, err := op.mempool.consensus.Consensus().GetVirtualDAAScore()
	if err != nil {
		return err
	}

	if virtualDAAScore-op.lastExpireScan < op.mempool.config.OrphanExpireScanIntervalDAAScore {
		return nil
	}

	for _, orphanTransaction := range op.allOrphans {
		// Never expire high priority transactions
		if orphanTransaction.IsHighPriority() {
			continue
		}

		// Remove all transactions whose addedAtDAAScore is older then TransactionExpireIntervalDAAScore
		if virtualDAAScore-orphanTransaction.AddedAtDAAScore() > op.mempool.config.OrphanExpireIntervalDAAScore {
			err = op.removeOrphan(orphanTransaction.TransactionID(), false)
			if err != nil {
				return err
			}
		}
	}

	op.lastExpireScan = virtualDAAScore
	return nil
}

func (op *orphansPool) updateOrphansAfterTransactionRemoved(
	removedTransaction *model.MempoolTransaction, removeRedeemers bool) error {

	if removeRedeemers {
		return op.removeRedeemersOf(removedTransaction)
	}

	outpoint := externalapi.DomainOutpoint{TransactionID: *removedTransaction.TransactionID()}
	for i := range removedTransaction.Transaction().Outputs {
		outpoint.Index = uint32(i)
		if orphan, ok := op.orphansByPreviousOutpoint[outpoint]; ok {
			for _, input := range orphan.Transaction().Inputs {
				if input.PreviousOutpoint.TransactionID.Equal(removedTransaction.TransactionID()) {
					input.UTXOEntry = nil
				}
			}
		}
	}

	return nil
}

func (op *orphansPool) randomNonHighPriorityOrphan() *model.OrphanTransaction {
	for _, orphan := range op.allOrphans {
		if !orphan.IsHighPriority() {
			return orphan
		}
	}

	return nil
}
