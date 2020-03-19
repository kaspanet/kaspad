package database2

// DatabaseHandle defines the interface of a database
// that can begin transactions and close itself.
// Important: this is not part of the main Database interface
// because the Transaction interface includes it. Were we to
// merge DatabaseHandle with Database, implementors of the
// Transaction interface would be forced to implement the
// Begin and Close methods, which is undesirable.
type DatabaseHandle interface {
	// A handle to the database needs to be able to do
	// anything that the underlying the database can do.
	Database

	// Begin begins a new database transaction.
	Begin() (Transaction, error)

	// Close closes the database.
	Close() error
}
