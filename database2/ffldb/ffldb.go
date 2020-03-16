package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/flatfile"
	"github.com/kaspanet/kaspad/database2/ffldb/leveldb"
)

const blockStoreName = "block"

type Database struct {
	blockStore    *flatfile.FlatFileStore
	metadataStore *leveldb.LevelDB
}

func Open(path string) *Database {
	blockStore := flatfile.NewFlatFileStore(path, blockStoreName)
	metadataStore := leveldb.NewLevelDB(path)
	return &Database{
		blockStore:    blockStore,
		metadataStore: metadataStore,
	}
}

func (db *Database) Close() error {
	return nil
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
