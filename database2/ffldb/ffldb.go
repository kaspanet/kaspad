package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/bucket"
	"github.com/kaspanet/kaspad/database2/ffldb/flatfile"
	"github.com/kaspanet/kaspad/database2/ffldb/leveldb"
)

// Database defines the interface of the ffldb database.
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

	// AppendFlatData appends the given data to the flat
	// file store defined by storeName. This function
	// returns a serialized location handle that's meant
	// to be stored and later used when querying the data
	// that has just now been inserted.
	AppendFlatData(storeName string, data []byte) ([]byte, error)

	// RetrieveFlatData retrieves data from the flat file
	// stored defined by storeName using the given serialized
	// location handle. See AppendFlatData for further details.
	RetrieveFlatData(storeName string, location []byte) ([]byte, error)

	// CurrentFlatDataLocation returns the serialized
	// location handle to the current location within
	// the flat file store defined storeName. It is mainly
	// to be used to rollback flat file stores in case
	// of data incongruency.
	CurrentFlatDataLocation(storeName string) []byte

	// RollbackFlatData truncates the flat file store defined
	// by the given storeName to the location defined by the
	// given serialized location handle.
	RollbackFlatData(storeName string, location []byte) error
}

// FFLDB is a database that's both LevelDB and a flat-file
// database.
type FFLDB struct {
	ffdb *flatfile.FFDB
	ldb  *leveldb.LDB
}

// Open opens a new ffldb with the given path.
func Open(path string) (*FFLDB, error) {
	ffdb := flatfile.NewFlatFileDatabase(path)
	ldb, err := leveldb.NewLevelDB(path)
	if err != nil {
		return nil, err
	}

	db := &FFLDB{
		ffdb: ffdb,
		ldb:  ldb,
	}
	return db, nil
}

// Close closes the database.
func (db *FFLDB) Close() error {
	err := db.ffdb.Close()
	if err != nil {
		return err
	}
	return db.ldb.Close()
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the Database interface.
func (db *FFLDB) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value)
}

// Get gets the value for the given key. It returns an
// error if the given key does not exist.
// This method is part of the Database interface.
func (db *FFLDB) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the Database interface.
func (db *FFLDB) Has(key []byte) (bool, error) {
	return db.ldb.Has(key)
}

// AppendFlatData appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the Database interface.
func (db *FFLDB) AppendFlatData(storeName string, data []byte) ([]byte, error) {
	return db.ffdb.Write(storeName, data)
}

// RetrieveFlatData retrieves data from the flat file
// stored defined by storeName using the given serialized
// location handle. See AppendFlatData for further details.
// This method is part of the Database interface.
func (db *FFLDB) RetrieveFlatData(storeName string, location []byte) ([]byte, error) {
	return db.ffdb.Read(storeName, location)
}

// CurrentFlatDataLocation returns the serialized
// location handle to the current location within
// the flat file store defined storeName. It is mainly
// to be used to rollback flat file stores in case
// of data incongruency.
// This method is part of the Database interface.
func (db *FFLDB) CurrentFlatDataLocation(storeName string) []byte {
	return db.ffdb.CurrentLocation(storeName)
}

// RollbackFlatData truncates the flat file store defined
// by the given storeName to the location defined by the
// given serialized location handle.
// This method is part of the Database interface.
func (db *FFLDB) RollbackFlatData(storeName string, location []byte) error {
	return db.ffdb.Rollback(storeName, location)
}

// Transaction is an ffldb transaction.
type Transaction struct {
	ldbTx *leveldb.LDBTransaction
	ffdb  *flatfile.FFDB
}

// Begin begins a new ffldb transaction.
func (db *FFLDB) Begin() (*Transaction, error) {
	ldbTx, err := db.ldb.Begin()
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		ldbTx: ldbTx,
		ffdb:  db.ffdb,
	}
	return transaction, nil
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the Database interface.
func (tx *Transaction) Put(key []byte, value []byte) error {
	return tx.ldbTx.Put(key, value)
}

// Get gets the value for the given key. It returns an
// error if the given key does not exist.
// This method is part of the Database interface.
func (tx *Transaction) Get(key []byte) ([]byte, error) {
	return tx.ldbTx.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the Database interface.
func (tx *Transaction) Has(key []byte) (bool, error) {
	return tx.ldbTx.Has(key)
}

// AppendFlatData appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the Database interface.
func (tx *Transaction) AppendFlatData(storeName string, data []byte) ([]byte, error) {
	return tx.ffdb.Write(storeName, data)
}

// RetrieveFlatData retrieves data from the flat file
// stored defined by storeName using the given serialized
// location handle. See AppendFlatData for further details.
// This method is part of the Database interface.
func (tx *Transaction) RetrieveFlatData(storeName string, location []byte) ([]byte, error) {
	return tx.ffdb.Read(storeName, location)
}

// CurrentFlatDataLocation returns the serialized
// location handle to the current location within
// the flat file store defined storeName. It is mainly
// to be used to rollback flat file stores in case
// of data incongruency.
// This method is part of the Database interface.
func (tx *Transaction) CurrentFlatDataLocation(storeName string) []byte {
	return tx.ffdb.CurrentLocation(storeName)
}

// RollbackFlatData truncates the flat file store defined
// by the given storeName to the location defined by the
// given serialized location handle.
// This method is part of the Database interface.
func (tx *Transaction) RollbackFlatData(storeName string, location []byte) error {
	return tx.ffdb.Rollback(storeName, location)
}

// Rollback rolls back whatever changes were made to the
// database within this transaction.
func (tx *Transaction) Rollback() error {
	return tx.ldbTx.Rollback()
}

// Commit commits whatever changes were made to the database
// within this transaction.
func (tx *Transaction) Commit() error {
	return tx.ldbTx.Commit()
}

// Key returns a key using the given key value and the
// given path of buckets.
// Example:
// * key: aaa
// * buckets: bbb, ccc
// * Result: bbb/ccc/aaa
func Key(key []byte, buckets ...[]byte) []byte {
	return bucket.BuildKey(key, buckets...)
}

// BucketPath returns a compound path using the given
// path of buckets.
// Example:
// * buckets: bbb, ccc
// * Result: bbb/ccc/
func BucketPath(buckets ...[]byte) []byte {
	return bucket.BuildBucketPath(buckets...)
}
