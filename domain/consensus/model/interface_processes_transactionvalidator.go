package model

import (
	"github.com/c4ei/kaspad/domain/consensus/model/externalapi"
)

// TransactionValidator exposes a set of validation classes, after which
// it's possible to determine whether a transaction is valid
type TransactionValidator interface {
	ValidateTransactionInIsolation(transaction *externalapi.DomainTransaction, povDAAScore uint64) error
	ValidateTransactionInContextIgnoringUTXO(stagingArea *StagingArea, tx *externalapi.DomainTransaction,
		povBlockHash *externalapi.DomainHash, povBlockPastMedianTime int64) error
	ValidateTransactionInContextAndPopulateFee(stagingArea *StagingArea,
		tx *externalapi.DomainTransaction, povBlockHash *externalapi.DomainHash) error
	PopulateMass(transaction *externalapi.DomainTransaction)
}
