package model

// TransactionValidator exposes a set of validation classes, after which
// it's possible to determine whether a transaction is valid
type TransactionValidator interface {
	ValidateTransactionAndPopulateWithConsensusData(transaction *DomainTransaction) error
}
