package ffldb

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/database2/ffldb/ff"
	"github.com/kaspanet/kaspad/database2/ffldb/ldb"
	"github.com/pkg/errors"
)

var (
	// flatFilesBucket keeps an index flat-file stores and their
	// current locations. Among other things, it is used to repair
	// the database in case a corruption occurs.
	flatFilesBucket = database2.MakeBucket([]byte("flat-files"))
)

// ffldb is a database utilizing LevelDB for key-value data and
// flat-files for raw data storage.
type ffldb struct {
	ffdb *ff.FlatFileDB
	ldb  *ldb.LevelDB
}

// Open opens a new ffldb with the given path.
func Open(path string) (database2.Database, error) {
	ffdb := ff.NewFlatFileDB(path)
	ldb, err := ldb.NewLevelDB(path)
	if err != nil {
		return nil, err
	}

	db := &ffldb{
		ffdb: ffdb,
		ldb:  ldb,
	}

	err = db.initialize()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the database.
// This method is part of the Database interface.
func (db *ffldb) Close() error {
	err := db.ffdb.Close()
	if err != nil {
		ldbCloseErr := db.ldb.Close()
		if ldbCloseErr != nil {
			return errors.Errorf("flat file db and leveldb both failed to close. "+
				"Errors: `%s`, `%s`", err, ldbCloseErr)
		}
		return err
	}
	return db.ldb.Close()
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the DataAccessor interface.
func (db *ffldb) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value)
}

// Get gets the value for the given key. It returns nil if
// the given key does not exist.
// This method is part of the DataAccessor interface.
func (db *ffldb) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the DataAccessor interface.
func (db *ffldb) Has(key []byte) (bool, error) {
	return db.ldb.Has(key)
}

// AppendToStore appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the DataAccessor interface.
func (db *ffldb) AppendToStore(storeName string, data []byte) ([]byte, error) {
	location, err := db.ffdb.Write(storeName, data)
	if err != nil {
		return nil, err
	}

	err = db.updateCurrentStoreLocation(storeName, location)
	if err != nil {
		return nil, err
	}

	return location, err
}

func (db *ffldb) updateCurrentStoreLocation(storeName string, location []byte) error {
	locationKey := flatFilesBucket.Key([]byte(storeName))
	return db.Put(locationKey, location)
}

// RetrieveFromStore retrieves data from the flat file
// stored defined by storeName using the given serialized
// location handle. See AppendToStore for further details.
// This method is part of the DataAccessor interface.
func (db *ffldb) RetrieveFromStore(storeName string, location []byte) ([]byte, error) {
	return db.ffdb.Read(storeName, location)
}

// CurrentStoreLocation returns the serialized
// location handle to the current location within
// the flat file store defined storeName. It is mainly
// to be used to rollback flat file stores in case
// of data incongruency.
// This method is part of the DataAccessor interface.
func (db *ffldb) CurrentStoreLocation(storeName string) []byte {
	return db.ffdb.CurrentLocation(storeName)
}

// RollbackStore truncates the flat file store defined
// by the given storeName to the location defined by the
// given serialized location handle.
// This method is part of the DataAccessor interface.
func (db *ffldb) RollbackStore(storeName string, location []byte) error {
	return db.ffdb.Rollback(storeName, location)
}

// Begin begins a new ffldb transaction.
// This method is part of the Database interface.
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

// Cursor begins a new cursor over the given bucket.
// This method is part of the Database interface.
func (db *ffldb) Cursor(bucket []byte) (database2.Cursor, error) {
	ldbCursor := db.ldb.Cursor(bucket)
	cursor := &cursor{
		ldbCursor: ldbCursor,
	}

	return cursor, nil
}
