package ldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

// LevelDB defines a thin wrapper around leveldb.
type LevelDB struct {
	ldb *leveldb.DB
}

// NewLevelDB opens a leveldb instance defined by the given path.
func NewLevelDB(path string) (*LevelDB, error) {
	// Open leveldb. If it doesn't exist, create it.
	ldb, err := leveldb.OpenFile(path, nil)

	// If the database is corrupted, attempt to recover.
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		log.Warnf("LevelDB corruption detected for path %s: %s",
			path, err)
		var err error
		ldb, err = leveldb.RecoverFile(path, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("LevelDB recovered from corruption for path %s",
			path)
	}

	// If the database cannot be opened for any other
	// reason, return the error as-is.
	if err != nil {
		return nil, err
	}

	db := &LevelDB{
		ldb: ldb,
	}
	return db, nil
}

// Close closes the leveldb instance.
func (db *LevelDB) Close() error {
	return db.ldb.Close()
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
func (db *LevelDB) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value, nil)
}

// Get gets the value for the given key. It returns an
// error if the given key does not exist.
func (db *LevelDB) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key, nil)
}

// Has returns true if the database does contains the
// given key.
func (db *LevelDB) Has(key []byte) (bool, error) {
	return db.ldb.Has(key, nil)
}
