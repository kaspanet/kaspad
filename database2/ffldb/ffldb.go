package ffldb

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/database2/ffldb/ff"
	"github.com/kaspanet/kaspad/database2/ffldb/ldb"
)

// ffldb is a database that's both LevelDB and a flat-file
// database.
type ffldb struct {
	ffdb *ff.FlatFileDB
	ldb  *ldb.LevelDB
}

// Open opens a new ffldb with the given path.
func Open(path string) (database2.Handle, error) {
	ffdb := ff.NewFlatFileDB(path)
	ldb, err := ldb.NewLevelDB(path)
	if err != nil {
		return nil, err
	}

	db := &ffldb{
		ffdb: ffdb,
		ldb:  ldb,
	}
	return db, nil
}

// Close closes the database.
// This method is part of the Handle interface.
func (db *ffldb) Close() error {
	err := db.ffdb.Close()
	if err != nil {
		return err
	}
	return db.ldb.Close()
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the Database interface.
func (db *ffldb) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value)
}

// Get gets the value for the given key. It returns an
// error if the given key does not exist.
// This method is part of the Database interface.
func (db *ffldb) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the Database interface.
func (db *ffldb) Has(key []byte) (bool, error) {
	return db.ldb.Has(key)
}

// AppendToStore appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the Database interface.
func (db *ffldb) AppendToStore(storeName string, data []byte) ([]byte, error) {
	return db.ffdb.Write(storeName, data)
}

// RetrieveFromStore retrieves data from the flat file
// stored defined by storeName using the given serialized
// location handle. See AppendToStore for further details.
// This method is part of the Database interface.
func (db *ffldb) RetrieveFromStore(storeName string, location []byte) ([]byte, error) {
	return db.ffdb.Read(storeName, location)
}

// CurrentFlatDataLocation returns the serialized
// location handle to the current location within
// the flat file store defined storeName. It is mainly
// to be used to rollback flat file stores in case
// of data incongruency.
// This method is part of the Database interface.
func (db *ffldb) CurrentFlatDataLocation(storeName string) []byte {
	return db.ffdb.CurrentLocation(storeName)
}

// RollbackFlatData truncates the flat file store defined
// by the given storeName to the location defined by the
// given serialized location handle.
// This method is part of the Database interface.
func (db *ffldb) RollbackFlatData(storeName string, location []byte) error {
	return db.ffdb.Rollback(storeName, location)
}

// Begin begins a new ffldb transaction.
// This method is part of the Handle interface.
func (db *ffldb) Begin() (database2.Transaction, error) {
	ldbTx, err := db.ldb.Begin()
	if err != nil {
		return nil, err
	}

	transaction := &transaction{
		ldbTx: ldbTx,
		ffdb:  db.ffdb,
	}
	return transaction, nil
}

// Key returns a key using the given key value and the
// given path of buckets.
// Example:
// * key: aaa
// * buckets: bbb, ccc
// * Result: bbb/ccc/aaa
func Key(key []byte, buckets ...[]byte) []byte {
	return ldb.BuildKey(key, buckets...)
}

// BucketPath returns a compound path using the given
// path of buckets.
// Example:
// * buckets: bbb, ccc
// * Result: bbb/ccc/
func BucketPath(buckets ...[]byte) []byte {
	return ldb.BuildBucketPath(buckets...)
}
