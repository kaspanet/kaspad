package ldb

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// LevelDBTransaction is a thin wrapper around native leveldb
// batches. It supports both get and put.
//
// Note that reads are done from the Database directly, so if another transaction changed the data,
// you will read the new data, and not the one from the time the transaction was opened/
//
// Note: As it's currently implemented, if one puts data into the transaction
// then it will not be available to get within the same transaction.
type LevelDBTransaction struct {
	db       *LevelDB
	batch    *leveldb.Batch
	isClosed bool
}

// Begin begins a new transaction.
func (db *LevelDB) Begin() (database.Transaction, error) {
	batch := new(leveldb.Batch)

	transaction := &LevelDBTransaction{
		db:       db,
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
	return errors.WithStack(tx.db.ldb.Write(tx.batch, nil))
}

// Rollback rolls back whatever changes were made to the
// database within this transaction.
func (tx *LevelDBTransaction) Rollback() error {
	if tx.isClosed {
		return errors.New("cannot rollback a closed transaction")
	}

	tx.isClosed = true
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
func (tx *LevelDBTransaction) Put(key *database.Key, value []byte) error {
	if tx.isClosed {
		return errors.New("cannot put into a closed transaction")
	}

	tx.batch.Put(key.Bytes(), value)
	return nil
}

// Get gets the value for the given key. It returns
// ErrNotFound if the given key does not exist.
func (tx *LevelDBTransaction) Get(key *database.Key) ([]byte, error) {
	if tx.isClosed {
		return nil, errors.New("cannot get from a closed transaction")
	}
	return tx.db.Get(key)
}

// Has returns true if the database does contains the
// given key.
func (tx *LevelDBTransaction) Has(key *database.Key) (bool, error) {
	if tx.isClosed {
		return false, errors.New("cannot has from a closed transaction")
	}
	return tx.db.Has(key)
}

// Delete deletes the value for the given key. Will not
// return an error if the key doesn't exist.
func (tx *LevelDBTransaction) Delete(key *database.Key) error {
	if tx.isClosed {
		return errors.New("cannot delete from a closed transaction")
	}

	tx.batch.Delete(key.Bytes())
	return nil
}

// Cursor begins a new cursor over the given bucket.
func (tx *LevelDBTransaction) Cursor(bucket *database.Bucket) (database.Cursor, error) {
	if tx.isClosed {
		return nil, errors.New("cannot open a cursor from a closed transaction")
	}

	return tx.db.Cursor(bucket)
}
