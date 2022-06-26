package libkaspawallet

import (
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TODO: Implement a better fee construction and estimation mechanism.

//FeePerInput is the current constant per input to pay for transactions.
const FeePerInput uint64 = 10000

//CalculateFees calculates the totalFee for a slice of transactions
func CalculateFees(transactions []*externalapi.DomainTransaction) (uint64, error) {

	var totalFee uint64

	for _, tx := range transactions {
		fee, err := CalculateFee(tx)
		if err != nil {
			return 0, err
		}
		totalFee += fee
	}

	return totalFee, nil
}

//CalculateFee calculates fee for a transaction
func CalculateFee(transaction *externalapi.DomainTransaction) (uint64, error) {

	var totalInputAmount uint64
	var totalOutputAmount uint64

	for _, input := range transaction.Inputs {
		totalInputAmount += input.UTXOEntry.Amount()
	}
	for _, output := range transaction.Outputs {
		totalOutputAmount += output.Value
	}

	if totalOutputAmount > totalInputAmount {
		return 0, errors.Errorf("Cannot Calculate negative fees")
	}

	return totalInputAmount - totalOutputAmount, nil
}
