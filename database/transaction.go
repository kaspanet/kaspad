package database

// Transaction defines the interface of a generic kaspad database
// transaction.
//
// Note: Transactions provide data consistency over the state of
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

	// RollbackUnlessClosed rolls back changes that were made to
	// the database within the transaction, unless the transaction
	// had already been closed using either Rollback or Commit.
	RollbackUnlessClosed() error
}
