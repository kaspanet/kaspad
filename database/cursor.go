package database

// Cursor iterates over database entries given some bucket.
type Cursor interface {
	// Next moves the iterator to the next key/value pair. It returns whether the
	// iterator is exhausted. Returns false if the cursor is closed.
	Next() bool

	// First moves the iterator to the first key/value pair. It returns false if
	// such a pair does not exist or if the cursor is closed.
	First() bool

	// Seek moves the iterator to the first key/value pair whose key is greater
	// than or equal to the given key. It returns ErrNotFound if such pair does not
	// exist.
	Seek(key []byte) error

	// Key returns the key of the current key/value pair, or ErrNotFound if done.
	// Note that the key is trimmed to not include the prefix the cursor was opened
	// with. The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to Next.
	Key() ([]byte, error)

	// Value returns the value of the current key/value pair, or ErrNotFound if done.
	// The caller should not modify the contents of the returned slice, and its
	// contents may change on the next call to Next.
	Value() ([]byte, error)

	// Close releases associated resources.
	Close() error
}
