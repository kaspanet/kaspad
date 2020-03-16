package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/flatfile"
)

const blockStoreName = "block"

type Database struct {
	blockStore *flatfile.FlatFileStore
}

func Open(path string) *Database {
	blockStore := flatfile.NewFlatFileStore(path, blockStoreName)
	return &Database{
		blockStore: blockStore,
	}
}

func (db *Database) Close() error {
	return db.Close()
}

func (db *Database) Begin() *Database {
	return db
}

func (db *Database) Rollback() error {
	return nil
}

func (db *Database) Commit() error {
	return nil
}

func (db *Database) Put(key string, value []byte) error {
	return nil
}

func (db *Database) Get(key string) ([]byte, error) {
	return nil, nil
}
