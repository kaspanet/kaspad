package ldb

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	ldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
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
	if _, corrupted := err.(*ldbErrors.ErrCorrupted); corrupted {
		log.Warnf("LevelDB corruption detected for path %s: %s",
			path, err)
		var recoverErr error
		ldb, recoverErr = leveldb.RecoverFile(path, nil)
		if recoverErr != nil {
			return nil, errors.Wrapf(err, "failed recovering from "+
				"database corruption: %s", recoverErr)
		}
		log.Warnf("LevelDB recovered from corruption for path %s",
			path)
	}

	// If the database cannot be opened for any other
	// reason, return the error as-is.
	if err != nil {
		return nil, errors.WithStack(err)
	}

	db := &LevelDB{
		ldb: ldb,
	}
	return db, nil
}

// Close closes the leveldb instance.
func (db *LevelDB) Close() error {
	err := db.ldb.Close()
	return errors.WithStack(err)
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
func (db *LevelDB) Put(key []byte, value []byte) error {
	err := db.ldb.Put(key, value, nil)
	return errors.WithStack(err)
}

// Get gets the value for the given key. It returns
// ErrNotFound if the given key does not exist.
func (db *LevelDB) Get(key []byte) ([]byte, error) {
	data, err := db.ldb.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, errors.Wrapf(database.ErrNotFound,
				"key %s not found", hex.EncodeToString(key))
		}
		return nil, errors.WithStack(err)
	}
	return data, nil
}

// Has returns true if the database does contains the
// given key.
func (db *LevelDB) Has(key []byte) (bool, error) {
	exists, err := db.ldb.Has(key, nil)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return exists, nil
}

// Delete deletes the value for the given key. Will not
// return an error if the key doesn't exist.
func (db *LevelDB) Delete(key []byte) error {
	err := db.ldb.Delete(key, nil)
	return errors.WithStack(err)
}
