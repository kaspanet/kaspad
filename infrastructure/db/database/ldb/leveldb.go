package ldb

import (
	"crypto/rand"
	"fmt"

	"github.com/kaspanet/kaspad/infrastructure/db/database"
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
	ldb, err := leveldb.OpenFile(path, Options())

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
func (db *LevelDB) Put(key *database.Key, value []byte) error {
	err := db.ldb.Put(key.Bytes(), value, nil)
	return errors.WithStack(err)
}

// Get gets the value for the given key. It returns
// ErrNotFound if the given key does not exist.
func (db *LevelDB) Get(key *database.Key) ([]byte, error) {
	data, err := db.ldb.Get(key.Bytes(), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, errors.Wrapf(database.ErrNotFound,
				"key %s not found", key)
		}
		return nil, errors.WithStack(err)
	}
	return data, nil
}

// Has returns true if the database does contains the
// given key.
func (db *LevelDB) Has(key *database.Key) (bool, error) {
	exists, err := db.ldb.Has(key.Bytes(), nil)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return exists, nil
}

// Delete deletes the value for the given key. Will not
// return an error if the key doesn't exist.
func (db *LevelDB) Delete(key *database.Key) error {
	err := db.ldb.Delete(key.Bytes(), nil)
	return errors.WithStack(err)
}

// AppendToStore appends the given data to the flat
// file store defined by storeName. This function
// returns a serialized location handle that's meant
// to be stored and later used when querying the data
// that has just now been inserted.
// This method is part of the DataAccessor interface.
func (db *LevelDB) AppendToStore(storeName string, data []byte) ([]byte, error) {
	return appendToStore(db, storeName, data)
}

// RetrieveFromStore retrieves data from the store defined by
// storeName using the given serialized location handle. It
// returns ErrNotFound if the location does not exist. See
// AppendToStore for further details.
// This method is part of the DataAccessor interface.
func (db *LevelDB) RetrieveFromStore(storeName string, location []byte) ([]byte, error) {
	return retrieveFromStore(db, storeName, location)
}

func randomKeySuffix() (string, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", key[0:4], key[4:6], key[6:8], key[8:10], key[10:]), nil
}

func appendToStore(accessor database.DataAccessor, storeName string, data []byte) ([]byte, error) {
	bucket := storeBucket(storeName)
	keySuffix, err := randomKeySuffix()
	if err != nil {
		return nil, err
	}

	key := bucket.Key([]byte(keySuffix))

	err = accessor.Put(key, data)
	if err != nil {
		return nil, err
	}

	return key.Bytes(), nil
}

var datastoreBucketKey = []byte("store")

func storeBucket(storeName string) *database.Bucket {
	return database.MakeBucket(datastoreBucketKey, []byte(storeName))
}

func retrieveFromStore(accessor database.DataAccessor, storeName string, location []byte) ([]byte, error) {
	key, err := database.KeyFromBytes(location)
	if err != nil {
		return nil, errors.Wrapf(database.ErrNotFound, "key %s is not valid", key)
	}

	return accessor.Get(key)
}
