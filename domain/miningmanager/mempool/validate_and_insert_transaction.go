package mempool

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	"github.com/kaspanet/kaspad/infrastructure/logger"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

func (mp *mempool) validateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool,
	allowOrphan bool) (acceptedTransactions []*externalapi.DomainTransaction, err error) {

	onEnd := logger.LogAndMeasureExecutionTime(log,
		fmt.Sprintf("validateAndInsertTransaction %s", consensushashing.TransactionID(transaction)))
	defer onEnd()

	if numOutsLessThanOneKas, isSpamming := isTXSpamming(transaction); isSpamming {
		log.Warnf("Rejected from mempool transaction %s with %d outputs with less than 1 KAS", consensushashing.TransactionID(transaction), numOutsLessThanOneKas)
		return nil, nil
	}

	// Populate mass in the beginning, it will be used in multiple places throughout the validation and insertion.
	mp.consensusReference.Consensus().PopulateMass(transaction)

	err = mp.validateTransactionPreUTXOEntry(transaction)
	if err != nil {
		return nil, err
	}

	parentsInPool, missingOutpoints, err := mp.fillInputsAndGetMissingParents(transaction)
	if err != nil {
		return nil, err
	}

	if len(missingOutpoints) > 0 {
		if !allowOrphan {
			str := fmt.Sprintf("Transaction %s is an orphan, where allowOrphan = false",
				consensushashing.TransactionID(transaction))
			return nil, transactionRuleError(RejectBadOrphan, str)
		}

		return nil, mp.orphansPool.maybeAddOrphan(transaction, isHighPriority)
	}

	err = mp.validateTransactionInContext(transaction)
	if err != nil {
		return nil, err
	}

	mempoolTransaction, err := mp.transactionsPool.addTransaction(transaction, parentsInPool, isHighPriority)
	if err != nil {
		return nil, err
	}

	acceptedOrphans, err := mp.orphansPool.processOrphansAfterAcceptedTransaction(mempoolTransaction.Transaction())
	if err != nil {
		return nil, err
	}

	acceptedTransactions = append([]*externalapi.DomainTransaction{transaction.Clone()}, acceptedOrphans...) //these pointer leave the mempool, hence we clone.

	err = mp.transactionsPool.limitTransactionCount()
	if err != nil {
		return nil, err
	}

	return acceptedTransactions, nil
}

func isTXSpamming(transaction *externalapi.DomainTransaction) (int, bool) {
	numOutsLessThanOneKas := 0
	for _, output := range transaction.Outputs {
		if output.Value < constants.SompiPerKaspa {
			numOutsLessThanOneKas += 1
		}
	}

	return numOutsLessThanOneKas, numOutsLessThanOneKas > len(transaction.Inputs)
}
