package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

func (mp *mempool) validateTransactionInIsolation(transaction *externalapi.DomainTransaction) error {
	transactionID := consensushashing.TransactionID(transaction)
	if _, ok := mp.transactionsPool.allTransactions[*transactionID]; ok {
		return transactionRuleError(RejectDuplicate,
			fmt.Sprintf("transaction %s is already in the mempool", transactionID))
	}

	if !mp.config.acceptNonStandard {
		if err := mp.checkTransactionStandard(transaction); err != nil {
			return err
		}
	}

	if err := mp.mempoolUTXOSet.checkDoubleSpends(transaction); err != nil {
		return err
	}

	return nil
}

func (mp *mempool) validateTransactionInContext(transaction *externalapi.DomainTransaction) error {
	if transaction.Mass > mp.config.maximumMassAcceptedByBlock {
		return transactionRuleError(RejectInvalid, fmt.Sprintf("transaction %s mass is %d which is "+
			"higher than the maxmimum of %d", consensushashing.TransactionID(transaction),
			transaction.Mass, mp.config.maximumMassAcceptedByBlock))
	}

	if !mp.config.acceptNonStandard {
		err := checkInputsStandard(transaction)
		if err != nil {
			return err
		}
	}

	return nil
}
