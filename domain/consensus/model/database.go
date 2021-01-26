package model

// DBCursor iterates over database entries given some bucket.
type DBCursor interface {
	// Next moves the iterator to the next key/value pair. It returns whether the
	// iterator is exhausted. Panics if the cursor is closed.
	Next() bool

	// First moves the iterator to the first key/value pair. It returns false if
	// such a pair does not exist. Panics if the cursor is closed.
	First() bool

	// Seek moves the iterator to the first key/value pair whose key is greater
	// than or equal to the given key. It returns ErrNotFound if such pair does not
	// exist.
	Seek(key DBKey) error

	// Key returns the key of the current key/value pair, or ErrNotFound if done.
	// The caller should not modify the contents of the returned key, and
	// its contents may change on the next call to Next.
	Key() (DBKey, error)

	// Value returns the value of the current key/value pair, or ErrNotFound if done.
	// The caller should not modify the contents of the returned slice, and its
	// contents may change on the next call to Next.
	Value() ([]byte, error)

	// Close releases associated resources.
	Close() error
}

// DBReader defines a proxy over domain data access
type DBReader interface {
	// Get gets the value for the given key. It returns
	// ErrNotFound if the given key does not exist.
	Get(key DBKey) ([]byte, error)

	// Has returns true if the database does contains the
	// given key.
	Has(key DBKey) (bool, error)

	// Cursor begins a new cursor over the given bucket.
	Cursor(bucket DBBucket) (DBCursor, error)
}

// DBWriter is an interface to write to the database
type DBWriter interface {
	DBReader

	// Put sets the value for the given key. It overwrites
	// any previous value for that key.
	Put(key DBKey, value []byte) error

	// Delete deletes the value for the given key. Will not
	// return an error if the key doesn't exist.
	Delete(key DBKey) error
}

// DBTransaction is a proxy over domain data
// access that requires an open database transaction
type DBTransaction interface {
	DBWriter

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

// DBManager defines the interface of a database that can begin
// transactions and read data.
type DBManager interface {
	DBWriter

	// Begin begins a new database transaction.
	Begin() (DBTransaction, error)
}

// DBKey is an interface for a database key
type DBKey interface {
	Bytes() []byte
	Bucket() DBBucket
	Suffix() []byte
}

// DBBucket is an interface for a database bucket
type DBBucket interface {
	Bucket(bucketBytes []byte) DBBucket
	Key(suffix []byte) DBKey
	Path() []byte
}
