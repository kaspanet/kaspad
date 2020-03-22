package database2

// Database defines the interface of a generic kaspad database.
type Database interface {
	// Put sets the value for the given key. It overwrites
	// any previous value for that key.
	Put(key []byte, value []byte) error

	// Get gets the value for the given key. It returns an
	// error if the given key does not exist.
	Get(key []byte) ([]byte, error)

	// Has returns true if the database does contains the
	// given key.
	Has(key []byte) (bool, error)

	// AppendToStore appends the given data to the store
	// defined by storeName. This function returns a serialized
	// location handle that's meant to be stored and later used
	// when querying the data that has just now been inserted.
	AppendToStore(storeName string, data []byte) ([]byte, error)

	// RetrieveFromStore retrieves data from the store defined by
	// storeName using the given serialized location handle. See
	// AppendToStore for further details.
	RetrieveFromStore(storeName string, location []byte) ([]byte, error)

	// CurrentStoreLocation returns the serialized
	// location handle to the current location within
	// the flat file store defined storeName. It is mainly
	// to be used to rollback flat file stores in case
	// of data incongruency.
	CurrentStoreLocation(storeName string) []byte

	// RollbackStore truncates the flat file store defined
	// by the given storeName to the location defined by the
	// given serialized location handle.
	RollbackStore(storeName string, location []byte) error
}
