package database2

// Transaction defines the interface of a generic kaspad database
// transaction.
// Note: transactions provide data consistency over the state of
// the database as it was when the transaction started. There is
// NO guarantee that if one puts data into the transaction then
// it will be available to get within the same transaction.
type Transaction interface {
	DataAccessor

	// Rollback rolls back whatever changes were made to the
	// database within this transaction.
	Rollback() error

	// Commit commits whatever changes were made to the database
	// within this transaction.
	Commit() error
}
