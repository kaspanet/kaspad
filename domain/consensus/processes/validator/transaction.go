package validator

import "github.com/kaspanet/kaspad/domain/consensus/model"

// ValidateTransactionAndCalculateFee validates the given transaction using
// the given utxoEntries. It also returns the transaction's fee
func (v *validator) ValidateTransactionAndCalculateFee(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) (fee uint64, err error) {
	err = v.checkTransactionInIsolation(transaction)
	if err != nil {
		return 0, err
	}

	return v.checkTransactionInContext()
}
