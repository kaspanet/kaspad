package externalapi

// TransactionStatus represents the confirmation state of the transaction.
type TransactionStatus byte

const (
	// StatusUnconfirmed indicates that the transaction is confirmed.
	StatusUnconfirmed TransactionStatus = iota

	// StatusConfirmed indicates that the transaction is confirmed.
	StatusConfirmed
)
