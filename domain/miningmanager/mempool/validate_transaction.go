package mempool

import (
	"fmt"
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/constants"

=======

	"github.com/zoomy-network/zoomyd/domain/consensus/utils/constants"

>>>>>>> Stashed changes
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/consensushashing"
)

func (mp *mempool) validateTransactionPreUTXOEntry(transaction *externalapi.DomainTransaction) error {
	err := mp.validateTransactionInIsolation(transaction)
	if err != nil {
		return err
	}

	if err := mp.mempoolUTXOSet.checkDoubleSpends(transaction); err != nil {
		return err
	}
	return nil
}

func (mp *mempool) validateTransactionInIsolation(transaction *externalapi.DomainTransaction) error {
	transactionID := consensushashing.TransactionID(transaction)
	if _, ok := mp.transactionsPool.allTransactions[*transactionID]; ok {
		return transactionRuleError(RejectDuplicate,
			fmt.Sprintf("transaction %s is already in the mempool", transactionID))
	}

	if !mp.config.AcceptNonStandard {
		if err := mp.checkTransactionStandardInIsolation(transaction); err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained. When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = RejectNonstandard
			}
			str := fmt.Sprintf("transaction %s is not standard: %s", transactionID, err)
			return transactionRuleError(rejectCode, str)
		}
	}

	return nil
}

func (mp *mempool) validateTransactionInContext(transaction *externalapi.DomainTransaction) error {
	hasCoinbaseInput := false
	for _, input := range transaction.Inputs {
		if input.UTXOEntry.IsCoinbase() {
			hasCoinbaseInput = true
			break
		}
	}

	numExtraOuts := len(transaction.Outputs) - len(transaction.Inputs)
	if !hasCoinbaseInput && numExtraOuts > 2 && transaction.Fee < uint64(numExtraOuts)*constants.SompiPerKaspa {
		log.Warnf("Rejected spam tx %s from mempool (%d outputs)", consensushashing.TransactionID(transaction), len(transaction.Outputs))
		return transactionRuleError(RejectSpamTx, fmt.Sprintf("Rejected spam tx %s from mempool", consensushashing.TransactionID(transaction)))
	}

	if !mp.config.AcceptNonStandard {
		err := mp.checkTransactionStandardInContext(transaction)
		if err != nil {
			// Attempt to extract a reject code from the error so
			// it can be retained. When not possible, fall back to
			// a non standard error.
			rejectCode, found := extractRejectCode(err)
			if !found {
				rejectCode = RejectNonstandard
			}
			str := fmt.Sprintf("transaction inputs %s are not standard: %s",
				consensushashing.TransactionID(transaction), err)
			return transactionRuleError(rejectCode, str)
		}
	}

	return nil
}
