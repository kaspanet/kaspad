package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/ff"
	"github.com/kaspanet/kaspad/database2/ffldb/ldb"
)

// transaction is an ffldb transaction.
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

// Get gets the value for the given key. It returns an
// error if the given key does not exist.
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

// AppendFlatData appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the Database interface.
func (tx *transaction) AppendFlatData(storeName string, data []byte) ([]byte, error) {
	return tx.ffdb.Write(storeName, data)
}

// RetrieveFlatData retrieves data from the flat file
// stored defined by storeName using the given serialized
// location handle. See AppendFlatData for further details.
// This method is part of the Database interface.
func (tx *transaction) RetrieveFlatData(storeName string, location []byte) ([]byte, error) {
	return tx.ffdb.Read(storeName, location)
}

// CurrentFlatDataLocation returns the serialized
// location handle to the current location within
// the flat file store defined storeName. It is mainly
// to be used to rollback flat file stores in case
// of data incongruency.
// This method is part of the Database interface.
func (tx *transaction) CurrentFlatDataLocation(storeName string) []byte {
	return tx.ffdb.CurrentLocation(storeName)
}

// RollbackFlatData truncates the flat file store defined
// by the given storeName to the location defined by the
// given serialized location handle.
// This method is part of the Database interface.
func (tx *transaction) RollbackFlatData(storeName string, location []byte) error {
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
