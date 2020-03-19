package database2

// Transaction defines the interface of a generic kaspad database
// transaction.
type Transaction interface {
	// A transaction needs to be able to do anything that the
	// underlying database can do.
	Database

	// Rollback rolls back whatever changes were made to the
	// database within this transaction.
	Rollback() error

	// Commit commits whatever changes were made to the database
	// within this transaction.
	Commit() error
}
