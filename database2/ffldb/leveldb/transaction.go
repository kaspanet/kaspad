package leveldb

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type LDBTransaction struct {
	ldb      *leveldb.DB
	snapshot *leveldb.Snapshot
	batch    *leveldb.Batch

	isClosed bool
}

func (db *LDB) Begin() (*LDBTransaction, error) {
	snapshot, err := db.ldb.GetSnapshot()
	if err != nil {
		return nil, err
	}
	batch := new(leveldb.Batch)

	transaction := &LDBTransaction{
		ldb:      db.ldb,
		snapshot: snapshot,
		batch:    batch,

		isClosed: false,
	}
	return transaction, nil
}

func (tx *LDBTransaction) Commit() error {
	if tx.isClosed {
		return errors.New("cannot commit a closed transaction")
	}

	tx.isClosed = true
	tx.snapshot.Release()
	return tx.ldb.Write(tx.batch, nil)
}

func (tx *LDBTransaction) Rollback() error {
	if tx.isClosed {
		return errors.New("cannot rollback a closed transaction")
	}

	tx.isClosed = true
	tx.snapshot.Release()
	tx.batch.Reset()
	return nil
}

func (tx *LDBTransaction) Put(key []byte, value []byte) error {
	if tx.isClosed {
		return errors.New("cannot put into a closed transaction")
	}

	tx.batch.Put(key, value)
	return nil
}

func (tx *LDBTransaction) Get(key []byte) ([]byte, error) {
	if tx.isClosed {
		return nil, errors.New("cannot get from a closed transaction")
	}

	return tx.snapshot.Get(key, nil)
}

func (tx *LDBTransaction) Has(key []byte) (bool, error) {
	if tx.isClosed {
		return false, errors.New("cannot has from a closed transaction")
	}

	return tx.snapshot.Has(key, nil)
}
