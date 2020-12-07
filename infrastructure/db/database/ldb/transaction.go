package ldb

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// LevelDBTransaction is a thin wrapper around native leveldb
// batches and snapshots. It supports both get and put.
//
// Snapshots provide a frozen view of the database at the moment
// the transaction begins. On the other hand, batches provide a
// mechanism to combine several database writes into one write,
// which seamlessly rolls back the database in case any individual
// write fails. Together the two forms a logic unit similar
// to what one might expect from a classic database transaction.
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

// Begin begins a new transaction.
func (db *LevelDB) Begin() (database.Transaction, error) {
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
	return errors.WithStack(tx.db.ldb.Write(tx.batch, &opt.WriteOptions{Sync: true}))
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

	data, err := tx.snapshot.Get(key.Bytes(), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, errors.Wrapf(database.ErrNotFound,
				"key %s not found", key)
		}
		return nil, errors.WithStack(err)
	}
	return data, nil
}

// Has returns true if the database does contains the
// given key.
func (tx *LevelDBTransaction) Has(key *database.Key) (bool, error) {
	if tx.isClosed {
		return false, errors.New("cannot has from a closed transaction")
	}

	res, err := tx.snapshot.Has(key.Bytes(), nil)
	return res, errors.WithStack(err)
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
