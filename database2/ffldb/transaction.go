package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/ff"
	"github.com/kaspanet/kaspad/database2/ffldb/ldb"
)

// transaction is an ffldb transaction.
// Note: transactions provide data consistency over the state of
// the database as it was when the transaction started. There is
// NO guarantee that if one puts data into the transaction then
// it will be available to get within the same transaction.
type transaction struct {
	ldbTx *ldb.LevelDBTransaction
	ffdb  *ff.FlatFileDB
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the Database interface.
func (tx *transaction) Put(key []byte, value []byte) error {
	return tx.ldbTx.Put(key, value)
}

// Get gets the value for the given key. It returns nil if
// the given key does not exist.
// This method is part of the Database interface.
func (tx *transaction) Get(key []byte) ([]byte, error) {
	return tx.ldbTx.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the Database interface.
func (tx *transaction) Has(key []byte) (bool, error) {
	return tx.ldbTx.Has(key)
}

// AppendToStore appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the Database interface.
func (tx *transaction) AppendToStore(storeName string, data []byte) ([]byte, error) {
	location, err := tx.ffdb.Write(storeName, data)
	if err != nil {
		return nil, err
	}

	err = updateCurrentStoreLocation(tx, storeName)
	if err != nil {
		return nil, err
	}

	return location, err
}

// RetrieveFromStore retrieves data from the flat file
// stored defined by storeName using the given serialized
// location handle. See AppendToStore for further details.
// This method is part of the Database interface.
func (tx *transaction) RetrieveFromStore(storeName string, location []byte) ([]byte, error) {
	return tx.ffdb.Read(storeName, location)
}

// CurrentStoreLocation returns the serialized
// location handle to the current location within
// the flat file store defined storeName. It is mainly
// to be used to rollback flat file stores in case
// of data incongruency.
// This method is part of the Database interface.
func (tx *transaction) CurrentStoreLocation(storeName string) ([]byte, error) {
	return tx.ffdb.CurrentLocation(storeName)
}

// RollbackStore truncates the flat file store defined
// by the given storeName to the location defined by the
// given serialized location handle.
// This method is part of the Database interface.
func (tx *transaction) RollbackStore(storeName string, location []byte) error {
	return tx.ffdb.Rollback(storeName, location)
}

// Rollback rolls back whatever changes were made to the
// database within this transaction.
// This method is part of the Transaction interface.
func (tx *transaction) Rollback() error {
	return tx.ldbTx.Rollback()
}

// Commit commits whatever changes were made to the database
// within this transaction.
// This method is part of the Transaction interface.
func (tx *transaction) Commit() error {
	return tx.ldbTx.Commit()
}
