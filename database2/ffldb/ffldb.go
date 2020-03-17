package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/bucket"
	"github.com/kaspanet/kaspad/database2/ffldb/flatfile"
	"github.com/kaspanet/kaspad/database2/ffldb/leveldb"
)

const blockStoreName = "block"
const metadataStoreName = "metadata"

type Database struct {
	blockStore    *flatfile.FlatFileStore
	metadataStore *leveldb.LevelDB
}

func Open(path string) (*Database, error) {
	blockStore := flatfile.NewFlatFileStore(path, blockStoreName)
	metadataStore, err := leveldb.NewLevelDB(path, metadataStoreName)
	if err != nil {
		return nil, err
	}

	db := &Database{
		blockStore:    blockStore,
		metadataStore: metadataStore,
	}
	return db, nil
}

func (db *Database) Close() error {
	return db.metadataStore.Close()
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

func (db *Database) Key(buckets ...[]byte) []byte {
	return bucket.BuildKey(buckets...)
}

func (db *Database) BucketKey(buckets ...[]byte) []byte {
	return bucket.BuildBucketKey(buckets...)
}

func (db *Database) Put(key string, value []byte) error {
	return nil
}

func (db *Database) Get(key string) ([]byte, error) {
	return nil, nil
}
