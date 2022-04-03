package database

// Database defines the interface of a database that can begin
// transactions and close itself.
//
// Important: This is not part of the DataAccessor interface
// because the Transaction interface includes it. Were we to
// merge Database with DataAccessor, implementors of the
// Transaction interface would be forced to implement methods
// such as Begin and Close, which is undesirable.
type Database interface {
	DataAccessor

	// Begin begins a new database transaction.
	Begin() (Transaction, error)

	// Compact compacts the database instance.
	Compact() error

	// Close closes the database.
	Close() error
}
