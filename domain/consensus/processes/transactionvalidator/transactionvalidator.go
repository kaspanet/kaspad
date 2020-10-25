package transactionvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// transactionValidator exposes a set of validation classes, after which
// it's possible to determine whether either a transaction is valid
type transactionValidator struct {
}

// New instantiates a new TransactionValidator
func New() model.TransactionValidator {
	return &transactionValidator{}
}

// ValidateTransactionAndCalculateFee validates the given transaction
// and populates it with any missing consensus data
func (v *transactionValidator) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	panic("implement me")
}
