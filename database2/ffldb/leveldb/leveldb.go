package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"path/filepath"
)

type LevelDB struct {
	ldb *leveldb.DB
}

func NewLevelDB(path string, storeName string) (*LevelDB, error) {
	dbPath := filepath.Join(path, storeName)

	// Open leveldb. If it doesn't exist, create it.
	ldb, err := leveldb.OpenFile(dbPath, nil)

	// If the database is corrupted, attempt to recover.
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		log.Warnf("LevelDB corruption detected for path %s: %s",
			dbPath, err)
		var err error
		ldb, err = leveldb.RecoverFile(dbPath, nil)
		if err != nil {
			return nil, err
		}
		log.Warnf("LevelDB recovered from corruption for path %s",
			dbPath)
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

func (db *LevelDB) Close() error {
	return db.ldb.Close()
}
