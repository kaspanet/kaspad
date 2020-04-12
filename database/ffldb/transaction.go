package ffldb

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/database/ffldb/ff"
	"github.com/kaspanet/kaspad/database/ffldb/ldb"
)

// transaction is an ffldb transaction.
//
// Note: Transactions provide data consistency over the state of
// the database as it was when the transaction started. There is
// NO guarantee that if one puts data into the transaction then
// it will be available to get within the same transaction.
type transaction struct {
	ldbTx *ldb.LevelDBTransaction
	ffdb  *ff.FlatFileDB
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the DataAccessor interface.
func (tx *transaction) Put(key *database.Key, value []byte) error {
	return tx.ldbTx.Put(key, value)
}

// Get gets the value for the given key. It returns
// ErrNotFound if the given key does not exist.
// This method is part of the DataAccessor interface.
func (tx *transaction) Get(key *database.Key) ([]byte, error) {
	return tx.ldbTx.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the DataAccessor interface.
func (tx *transaction) Has(key *database.Key) (bool, error) {
	return tx.ldbTx.Has(key)
}

// Delete deletes the value for the given key. Will not
// return an error if the key doesn't exist.
// This method is part of the DataAccessor interface.
func (tx *transaction) Delete(key *database.Key) error {
	return tx.ldbTx.Delete(key)
}

// AppendToStore appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the DataAccessor interface.
func (tx *transaction) AppendToStore(storeName string, data []byte) ([]byte, error) {
	return appendToStore(tx, tx.ffdb, storeName, data)
}

// RetrieveFromStore retrieves data from the store defined by
// storeName using the given serialized location handle. It
// returns ErrNotFound if the location does not exist. See
// AppendToStore for further details.
// This method is part of the DataAccessor interface.
func (tx *transaction) RetrieveFromStore(storeName string, location []byte) ([]byte, error) {
	return tx.ffdb.Read(storeName, location)
}

// Cursor begins a new cursor over the given bucket.
// This method is part of the DataAccessor interface.
func (tx *transaction) Cursor(bucket *database.Bucket) (database.Cursor, error) {
	return tx.ldbTx.Cursor(bucket)
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

// RollbackUnlessClosed rolls back changes that were made to
// the database within the transaction, unless the transaction
// had already been closed using either Rollback or Commit.
func (tx *transaction) RollbackUnlessClosed() error {
	return tx.ldbTx.RollbackUnlessClosed()
}
