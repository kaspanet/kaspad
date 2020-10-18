package validator

import "github.com/kaspanet/kaspad/domain/consensus/model"

// ValidateTransactionAndCalculateFee validates the given transaction using
// the given utxoEntries. It also returns the transaction's fee
func (v *validator) ValidateTransactionAndCalculateFee(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) (fee uint64, err error) {
	return 0, nil
}
