package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
)

// TransactionValidator exposes a set of validation classes, after which
// it's possible to determine whether a transaction is valid
type TransactionValidator interface {
	ValidateTransactionInIsolation(transaction *externalapi.DomainTransaction) error
	ValidateTransactionInContextAndPopulateMassAndFee(tx *externalapi.DomainTransaction,
		povTransactionHash *externalapi.DomainHash, selectedParentMedianTime int64) error
}

// TestTransactionValidator adds to the main TransactionValidator methods required by tests
type TestTransactionValidator interface {
	TransactionValidator
	SigCache() *txscript.SigCache
	SetSigCache(sigCache *txscript.SigCache)
}
