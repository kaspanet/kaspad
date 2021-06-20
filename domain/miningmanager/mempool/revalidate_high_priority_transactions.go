package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/pkg/errors"
)

func (mp *mempool) revalidateHighPriorityTransactions() ([]*externalapi.DomainTransaction, error) {
	validTransactions := []*externalapi.DomainTransaction{}

	for _, transaction := range mp.transactionsPool.highPriorityTransactions {
		isValid, err := mp.revalidateTransaction(transaction)
		if err != nil {
			return nil, err
		}
		if !isValid {
			continue
		}

		validTransactions = append(validTransactions, transaction.Transaction())
	}

	return validTransactions, nil
}

func (mp *mempool) revalidateTransaction(transaction *model.MempoolTransaction) (isValid bool, err error) {
	clearInputs(transaction)
	err = mp.mempoolUTXOSet.checkDoubleSpends(transaction.Transaction())
	if err != nil {
		if errors.As(err, &RuleError{}) {
			err := mp.removeTransaction(transaction.TransactionID(), true)
			if err != nil {
				return false, err
			}
			return false, nil
		}
		return false, err
	}

	_, missingParents, err := mp.fillInputsAndGetMissingParents(transaction.Transaction())
	if err != nil {
		return false, err
	}
	if len(missingParents) > 0 {
		err := mp.removeTransaction(transaction.TransactionID(), true)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func clearInputs(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction().Inputs {
		input.UTXOEntry = nil
	}
}
