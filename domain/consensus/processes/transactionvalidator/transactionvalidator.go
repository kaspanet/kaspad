package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// transactionValidator exposes a set of validation classes, after which
// it's possible to determine whether either a transaction is valid
type transactionValidator struct {
}

// New instantiates a new TransactionValidator
func New() model.TransactionValidator {
	return &transactionValidator{}
}

// ValidateTransactionAndCalculateFee validates the given transaction using
// the given utxoEntries. It also returns the transaction's fee
func (v *transactionValidator) ValidateTransactionAndCalculateFee(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) (fee uint64, err error) {
	return 0, nil
}
