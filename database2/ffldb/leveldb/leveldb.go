package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type LDB struct {
	ldb *leveldb.DB
}

func NewLevelDB(path string) (*LDB, error) {
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

	db := &LDB{
		ldb: ldb,
	}
	return db, nil
}

func (db *LDB) Close() error {
	return db.ldb.Close()
}

func (db *LDB) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value, nil)
}

func (db *LDB) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key, nil)
}

func (db *LDB) Has(key []byte) (bool, error) {
	return db.ldb.Has(key, nil)
}
