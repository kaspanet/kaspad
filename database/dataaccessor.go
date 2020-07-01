package database

// DataAccessor defines the common interface by which data gets
// accessed in a generic kaspad database.
type DataAccessor interface {
	// Put sets the value for the given key. It overwrites
	// any previous value for that key.
	Put(key *Key, value []byte) error

	// Get gets the value for the given key. It returns
	// ErrNotFound if the given key does not exist.
	Get(key *Key) ([]byte, error)

	// Has returns true if the database does contains the
	// given key.
	Has(key *Key) (bool, error)

	// Delete deletes the value for the given key. Will not
	// return an error if the key doesn't exist.
	Delete(key *Key) error

	// AppendToStore appends the given data to the store
	// defined by storeName. This function returns a serialized
	// location handle that's meant to be stored and later used
	// when querying the data that has just now been inserted.
	AppendToStore(storeName string, data []byte) ([]byte, error)

	// RetrieveFromStore retrieves data from the store defined by
	// storeName using the given serialized location handle. It
	// returns ErrNotFound if the location does not exist. See
	// AppendToStore for further details.
	RetrieveFromStore(storeName string, location []byte) ([]byte, error)

	// Cursor begins a new cursor over the given bucket.
	Cursor(bucket *Bucket) (Cursor, error)
}
