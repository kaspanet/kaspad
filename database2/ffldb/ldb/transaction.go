package ldb

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBTransaction is a thin wrapper around native leveldb
// batches and snapshots. It supports both get and put.
//
// Note: Transactions provide data consistency over the state of
// the database as it was when the transaction started. There is
// NO guarantee that if one puts data into the transaction then
// it will be available to get within the same transaction.
type LevelDBTransaction struct {
	ldb      *leveldb.DB
	snapshot *leveldb.Snapshot
	batch    *leveldb.Batch
	isClosed bool
}

// Begin begins a new transaction.
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

// Commit commits whatever changes were made to the database
// within this transaction.
func (tx *LevelDBTransaction) Commit() error {
	if tx.isClosed {
		return errors.New("cannot commit a closed transaction")
	}

	tx.isClosed = true
	tx.snapshot.Release()
	return tx.ldb.Write(tx.batch, nil)
}

// Rollback rolls back whatever changes were made to the
// database within this transaction.
func (tx *LevelDBTransaction) Rollback() error {
	if tx.isClosed {
		return errors.New("cannot rollback a closed transaction")
	}

	tx.isClosed = true
	tx.snapshot.Release()
	tx.batch.Reset()
	return nil
}

// Put sets the value for the given key. It overwrites
// any previous value for that key.
func (tx *LevelDBTransaction) Put(key []byte, value []byte) error {
	if tx.isClosed {
		return errors.New("cannot put into a closed transaction")
	}

	tx.batch.Put(key, value)
	return nil
}

// Get gets the value for the given key. It returns nil if
// the given key does not exist.
func (tx *LevelDBTransaction) Get(key []byte) ([]byte, error) {
	if tx.isClosed {
		return nil, errors.New("cannot get from a closed transaction")
	}

	data, err := tx.snapshot.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// Has returns true if the database does contains the
// given key.
func (tx *LevelDBTransaction) Has(key []byte) (bool, error) {
	if tx.isClosed {
		return false, errors.New("cannot has from a closed transaction")
	}

	return tx.snapshot.Has(key, nil)
}
