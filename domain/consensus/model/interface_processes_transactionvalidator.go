package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// TransactionValidator exposes a set of validation classes, after which
// it's possible to determine whether a transaction is valid
type TransactionValidator interface {
	ValidateTransactionInIsolation(transaction *externalapi.DomainTransaction) error
	ValidateTransactionInContextIgnoringUTXO(stagingArea *StagingArea, tx *externalapi.DomainTransaction,
		povBlockHash *externalapi.DomainHash) error
	ValidateTransactionInContextAndPopulateFee(stagingArea *StagingArea,
		tx *externalapi.DomainTransaction, povBlockHash *externalapi.DomainHash) error
	PopulateMass(transaction *externalapi.DomainTransaction)
}
