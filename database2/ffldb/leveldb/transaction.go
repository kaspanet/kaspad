package leveldb

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDBTransaction struct {
	ldb      *leveldb.DB
	snapshot *leveldb.Snapshot
	batch    *leveldb.Batch

	isClosed bool
}

func (db *LevelDB) Begin() (*LevelDBTransaction, error) {
	snapshot, err := db.ldb.GetSnapshot()
	if err != nil {
		return nil, err
	}
	batch := new(leveldb.Batch)

	transaction := &LevelDBTransaction{
		ldb:      db.ldb,
		snapshot: snapshot,
		batch:    batch,

		isClosed: false,
	}
	return transaction, nil
}

func (tx *LevelDBTransaction) Commit() error {
	if tx.isClosed {
		return errors.New("cannot commit a closed transaction")
	}

	tx.isClosed = true
	tx.snapshot.Release()
	return tx.ldb.Write(tx.batch, nil)
}

func (tx *LevelDBTransaction) Rollback() error {
	if tx.isClosed {
		return errors.New("cannot rollback a closed transaction")
	}

	tx.isClosed = true
	tx.snapshot.Release()
	tx.batch.Reset()
	return nil
}

func (tx *LevelDBTransaction) Put(key []byte, value []byte) error {
	if tx.isClosed {
		return errors.New("cannot put into a closed transaction")
	}

	tx.batch.Put(key, value)
	return nil
}

func (tx *LevelDBTransaction) Get(key []byte) ([]byte, error) {
	if tx.isClosed {
		return nil, errors.New("cannot get from a closed transaction")
	}

	return tx.snapshot.Get(key, nil)
}
