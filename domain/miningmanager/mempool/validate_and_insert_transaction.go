package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

func (mp *mempool) validateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool,
	allowOrphan bool) (acceptedTransactions []*externalapi.DomainTransaction, err error) {

	err = mp.validateTransactionInIsolation(transaction)
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

	acceptedTransactions = make([]*externalapi.DomainTransaction, 0, len(acceptedOrphans)+1)
	acceptedTransactions = append(acceptedTransactions, transaction)
	for _, acceptedOrphan := range acceptedOrphans {
		acceptedTransactions = append(acceptedTransactions, acceptedOrphan)
	}

	mp.transactionsPool.limitTransactionCount()

	return acceptedTransactions, nil
}
