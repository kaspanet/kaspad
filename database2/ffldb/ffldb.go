package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/bucket"
	"github.com/kaspanet/kaspad/database2/ffldb/flatfile"
	"github.com/kaspanet/kaspad/database2/ffldb/leveldb"
)

type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	AppendFlatData(storeName string, data []byte) ([]byte, error)
	RetrieveFlatData(storeName string, location []byte) ([]byte, error)
	RollbackFlatData(storeName string, location []byte) error
}

type FFLDB struct {
	ffdb *flatfile.FlatFileDB
	ldb  *leveldb.LevelDB
}

func Open(path string) (*FFLDB, error) {
	ffdb := flatfile.NewFlatFileDB(path)
	ldb, err := leveldb.NewLevelDB(path)
	if err != nil {
		return nil, err
	}

	db := &FFLDB{
		ffdb: ffdb,
		ldb:  ldb,
	}
	return db, nil
}

func (db *FFLDB) Close() error {
	return db.ldb.Close()
}

func (db *FFLDB) Put(key []byte, value []byte) error {
	return db.ldb.Put(key, value)
}

func (db *FFLDB) Get(key []byte) ([]byte, error) {
	return db.ldb.Get(key)
}

func (db *FFLDB) AppendFlatData(storeName string, data []byte) ([]byte, error) {
	return db.ffdb.Write(storeName, data)
}

func (db *FFLDB) RetrieveFlatData(storeName string, location []byte) ([]byte, error) {
	return db.ffdb.Read(storeName, location)
}

func (db *FFLDB) RollbackFlatData(storeName string, location []byte) error {
	return db.ffdb.Rollback(storeName, location)
}

type Transaction struct {
	ldbTx *leveldb.LevelDBTransaction
	ffdb  *flatfile.FlatFileDB
}

func (db *FFLDB) Begin() (*Transaction, error) {
	ldbTx, err := db.ldb.Begin()
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		ldbTx: ldbTx,
		ffdb:  db.ffdb,
	}
	return transaction, nil
}

func (tx *Transaction) Put(key []byte, value []byte) error {
	return tx.ldbTx.Put(key, value)
}

func (tx *Transaction) Get(key []byte) ([]byte, error) {
	return tx.ldbTx.Get(key)
}

func (tx *Transaction) AppendFlatData(storeName string, data []byte) ([]byte, error) {
	return tx.ffdb.Write(storeName, data)
}

func (tx *Transaction) RetrieveFlatData(storeName string, location []byte) ([]byte, error) {
	return tx.ffdb.Read(storeName, location)
}

func (tx *Transaction) RollbackFlatData(storeName string, location []byte) error {
	return tx.ffdb.Rollback(storeName, location)
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
