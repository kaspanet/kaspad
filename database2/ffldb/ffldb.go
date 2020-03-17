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
	AppendBlock(data []byte) ([]byte, error)
	RetrieveBlock(location []byte) ([]byte, error)
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

func (db *FFLDB) AppendBlock(data []byte) ([]byte, error) {
	return appendBlock(db.blockStore, data)
}

func (db *FFLDB) RetrieveBlock(serializedLocation []byte) ([]byte, error) {
	return retrieveBlock(db.blockStore, serializedLocation)
}

type Transaction struct {
	ldbTx      *leveldb.LevelDBTransaction
	blockStore *flatfile.FlatFileStore
}

func (db *FFLDB) Begin() (*Transaction, error) {
	ldbTx, err := db.metadataStore.Begin()
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		ldbTx:      ldbTx,
		blockStore: db.blockStore,
	}
	return transaction, nil
}

func (tx *Transaction) Put(key []byte, value []byte) error {
	return tx.ldbTx.Put(key, value)
}

func (tx *Transaction) Get(key []byte) ([]byte, error) {
	return tx.ldbTx.Get(key)
}

func (tx *Transaction) AppendBlock(data []byte) ([]byte, error) {
	return appendBlock(tx.blockStore, data)
}

func (tx *Transaction) RetrieveBlock(serializedLocation []byte) ([]byte, error) {
	return retrieveBlock(tx.blockStore, serializedLocation)
}

func (tx *Transaction) Rollback() error {
	return tx.ldbTx.Rollback()
}

func (tx *Transaction) Commit() error {
	return tx.ldbTx.Commit()
}

func appendBlock(blockStore *flatfile.FlatFileStore, data []byte) ([]byte, error) {
	location, err := blockStore.Write(data)
	if err != nil {
		return nil, err
	}
	return flatfile.SerializeLocation(location), nil
}

func retrieveBlock(blockStore *flatfile.FlatFileStore, serializedLocation []byte) ([]byte, error) {
	location, err := flatfile.DeserializeLocation(serializedLocation)
	if err != nil {
		return nil, err
	}
	return blockStore.Read(location)
}

func Key(buckets ...[]byte) []byte {
	return bucket.BuildKey(buckets...)
}

func BucketKey(buckets ...[]byte) []byte {
	return bucket.BuildBucketKey(buckets...)
}
