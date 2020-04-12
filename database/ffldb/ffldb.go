package ffldb

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/database/ffldb/ff"
	"github.com/kaspanet/kaspad/database/ffldb/ldb"
	"github.com/pkg/errors"
)

var (
	// flatFilesBucket keeps an index flat-file stores and their
	// current locations. Among other things, it is used to repair
	// the database in case a corruption occurs.
	flatFilesBucket = database.MakeBucket([]byte("flat-files"))
)

// ffldb is a database utilizing LevelDB for key-value data and
// flat-files for raw data storage.
type ffldb struct {
	flatFileDB *ff.FlatFileDB
	levelDB    *ldb.LevelDB
}

// Open opens a new ffldb with the given path.
func Open(path string) (database.Database, error) {
	flatFileDB := ff.NewFlatFileDB(path)
	levelDB, err := ldb.NewLevelDB(path)
	if err != nil {
		return nil, err
	}

	db := &ffldb{
		flatFileDB: flatFileDB,
		levelDB:    levelDB,
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
	err := db.flatFileDB.Close()
	if err != nil {
		ldbCloseErr := db.levelDB.Close()
		if ldbCloseErr != nil {
			return errors.Wrapf(err, "err occurred during leveldb close: %s", ldbCloseErr)
		}
		return err
	}
	return db.levelDB.Close()
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
// This method is part of the DataAccessor interface.
func (db *ffldb) Put(key *database.Key, value []byte) error {
	return db.levelDB.Put(key, value)
}

// Get gets the value for the given key. It returns
// ErrNotFound if the given key does not exist.
// This method is part of the DataAccessor interface.
func (db *ffldb) Get(key *database.Key) ([]byte, error) {
	return db.levelDB.Get(key)
}

// Has returns true if the database does contains the
// given key.
// This method is part of the DataAccessor interface.
func (db *ffldb) Has(key *database.Key) (bool, error) {
	return db.levelDB.Has(key)
}

// Delete deletes the value for the given key. Will not
// return an error if the key doesn't exist.
// This method is part of the DataAccessor interface.
func (db *ffldb) Delete(key *database.Key) error {
	return db.levelDB.Delete(key)
}

// AppendToStore appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the DataAccessor interface.
func (db *ffldb) AppendToStore(storeName string, data []byte) ([]byte, error) {
	return appendToStore(db, db.flatFileDB, storeName, data)
}

func appendToStore(accessor database.DataAccessor, ffdb *ff.FlatFileDB, storeName string, data []byte) ([]byte, error) {
	// Save a reference to the current location in case
	// we fail and need to rollback.
	previousLocation, err := ffdb.CurrentLocation(storeName)
	if err != nil {
		return nil, err
	}
	rollback := func() error {
		return ffdb.Rollback(storeName, previousLocation)
	}

	// Append the data to the store and rollback in case of an error.
	location, err := ffdb.Write(storeName, data)
	if err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return nil, errors.Wrapf(err, "error occurred during rollback: %s", rollbackErr)
		}
		return nil, err
	}

	// Get the new location. If this fails we won't be able to update
	// the current store location, in which case we roll back.
	currentLocation, err := ffdb.CurrentLocation(storeName)
	if err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return nil, errors.Wrapf(err, "error occurred during rollback: %s", rollbackErr)
		}
		return nil, err
	}

	// Set the current store location and roll back in case an error.
	err = setCurrentStoreLocation(accessor, storeName, currentLocation)
	if err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return nil, errors.Wrapf(err, "error occurred during rollback: %s", rollbackErr)
		}
		return nil, err
	}

	return location, err
}

func setCurrentStoreLocation(accessor database.DataAccessor, storeName string, location []byte) error {
	locationKey := flatFilesBucket.Key([]byte(storeName))
	return accessor.Put(locationKey, location)
}

// RetrieveFromStore retrieves data from the store defined by
// storeName using the given serialized location handle. It
// returns ErrNotFound if the location does not exist. See
// AppendToStore for further details.
// This method is part of the DataAccessor interface.
func (db *ffldb) RetrieveFromStore(storeName string, location []byte) ([]byte, error) {
	return db.flatFileDB.Read(storeName, location)
}

// Cursor begins a new cursor over the given bucket.
// This method is part of the DataAccessor interface.
func (db *ffldb) Cursor(bucket *database.Bucket) (database.Cursor, error) {
	ldbCursor := db.levelDB.Cursor(bucket)

	return ldbCursor, nil
}

// Begin begins a new ffldb transaction.
// This method is part of the Database interface.
func (db *ffldb) Begin() (database.Transaction, error) {
	ldbTx, err := db.levelDB.Begin()
	if err != nil {
		return nil, err
	}

	transaction := &transaction{
		ldbTx: ldbTx,
		ffdb:  db.flatFileDB,
	}
	return transaction, nil
}
