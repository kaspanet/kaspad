package ldb

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBTransaction is a thin wrapper around native leveldb
// batches and snapshots. It supports both get and put.
//
// Note: Transactions provide data consistency over the state of
// the database as it was when the transaction started. As it's
// currently implemented, if one puts data into the transaction
// then it will not be available to get within the same transaction.
type LevelDBTransaction struct {
	db       *LevelDB
	snapshot *leveldb.Snapshot
	batch    *leveldb.Batch
	isClosed bool
}

// Begin begins a new transaction. A transaction wraps two
// leveldb primitives: snapshots and batches. Snapshots provide
// a frozen view of the database at the moment the transaction
// begins. On the other hand, batches provide a mechanism to
// combine several database writes into one write, which
// seemlessly rolls back the database in case any individual
// write fails. Together the two forms a logic unit similar
// to what one might expect from a classic database transaction.
func (db *LevelDB) Begin() (*LevelDBTransaction, error) {
	snapshot, err := db.ldb.GetSnapshot()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	batch := new(leveldb.Batch)

	transaction := &LevelDBTransaction{
		db:       db,
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
	return tx.db.ldb.Write(tx.batch, nil)
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

// RollbackUnlessClosed rolls back changes that were made to
// the database within the transaction, unless the transaction
// had already been closed using either Rollback or Commit.
func (tx *LevelDBTransaction) RollbackUnlessClosed() error {
	if tx.isClosed {
		return nil
	}
	return tx.Rollback()
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
		return nil, errors.WithStack(err)
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

// Cursor begins a new cursor over the given bucket.
func (tx *LevelDBTransaction) Cursor(bucket []byte) (*LevelDBCursor, error) {
	if tx.isClosed {
		return nil, errors.New("cannot open a cursor from a closed transaction")
	}

	return tx.db.Cursor(bucket), nil
}
