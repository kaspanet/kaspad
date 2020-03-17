package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/bucket"
	"github.com/kaspanet/kaspad/database2/ffldb/flatfile"
	"github.com/kaspanet/kaspad/database2/ffldb/leveldb"
)

const blockStoreName = "block"
const metadataStoreName = "metadata"

type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
}

type FFLDB struct {
	blockStore    *flatfile.FlatFileStore
	metadataStore *leveldb.LevelDB
}

func Open(path string) (*FFLDB, error) {
	blockStore := flatfile.NewFlatFileStore(path, blockStoreName)
	metadataStore, err := leveldb.NewLevelDB(path, metadataStoreName)
	if err != nil {
		return nil, err
	}

	db := &FFLDB{
		blockStore:    blockStore,
		metadataStore: metadataStore,
	}
	return db, nil
}

func (db *FFLDB) Close() error {
	return db.metadataStore.Close()
}

func (db *FFLDB) Put(key []byte, value []byte) error {
	return db.metadataStore.Put(key, value)
}

func (db *FFLDB) Get(key []byte) ([]byte, error) {
	return db.metadataStore.Get(key)
}

type Transaction struct {
	ldbTx *leveldb.LevelDBTransaction
}

func (db *FFLDB) Begin() (*Transaction, error) {
	ldbTx, err := db.metadataStore.Begin()
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		ldbTx: ldbTx,
	}
	return transaction, nil
}

func (tx *Transaction) Put(key []byte, value []byte) error {
	return tx.ldbTx.Put(key, value)
}

func (tx *Transaction) Get(key []byte) ([]byte, error) {
	return tx.ldbTx.Get(key)
}

func (tx *Transaction) Rollback() error {
	return tx.ldbTx.Rollback()
}

func (tx *Transaction) Commit() error {
	return tx.ldbTx.Commit()
}

func Key(buckets ...[]byte) []byte {
	return bucket.BuildKey(buckets...)
}

func BucketKey(buckets ...[]byte) []byte {
	return bucket.BuildBucketKey(buckets...)
}
