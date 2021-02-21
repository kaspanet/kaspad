package ldb

import (
	"bytes"

	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBCursor is a thin wrapper around native leveldb iterators.
type LevelDBCursor struct {
	ldbIterator iterator.Iterator
	bucket      *database.Bucket

	isClosed bool
}

// Cursor begins a new cursor over the given prefix.
func (db *LevelDB) Cursor(bucket *database.Bucket) (database.Cursor, error) {
	ldbIterator := db.ldb.NewIterator(util.BytesPrefix(bucket.Path()), nil)

	return &LevelDBCursor{
		ldbIterator: ldbIterator,
		bucket:      bucket,
		isClosed:    false,
	}, nil
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted. Panics if the cursor is closed.
func (c *LevelDBCursor) Next() bool {
	if c.isClosed {
		panic("cannot call next on a closed cursor")
	}
	return c.ldbIterator.Next()
}

// First moves the iterator to the first key/value pair. It returns false if
// such a pair does not exist. Panics if the cursor is closed.
func (c *LevelDBCursor) First() bool {
	if c.isClosed {
		panic("cannot call first on a closed cursor")
	}
	return c.ldbIterator.First()
}

// Seek moves the iterator to the first key/value pair whose key is greater
// than or equal to the given key. It returns ErrNotFound if such pair does not
// exist.
func (c *LevelDBCursor) Seek(key *database.Key) error {
	if c.isClosed {
		return errors.New("cannot seek a closed cursor")
	}

	found := c.ldbIterator.Seek(key.Bytes())
	if !found {
		return errors.Wrapf(database.ErrNotFound, "key %s not found", key)
	}

	// Use c.ldbIterator.Key because c.Key removes the prefix from the key
	currentKey := c.ldbIterator.Key()
	if currentKey == nil || !bytes.Equal(currentKey, key.Bytes()) {
		return errors.Wrapf(database.ErrNotFound, "key %s not found", key)
	}

	return nil
}

// Key returns the key of the current key/value pair, or ErrNotFound if done.
// Note that the key is trimmed to not include the prefix the cursor was opened
// with. The caller should not modify the contents of the returned slice, and
// its contents may change on the next call to Next.
func (c *LevelDBCursor) Key() (*database.Key, error) {
	if c.isClosed {
		return nil, errors.New("cannot get the key of a closed cursor")
	}
	fullKeyPath := c.ldbIterator.Key()
	if fullKeyPath == nil {
		return nil, errors.Wrapf(database.ErrNotFound, "cannot get the "+
			"key of an exhausted cursor")
	}
	suffix := bytes.TrimPrefix(fullKeyPath, c.bucket.Path())
	return c.bucket.Key(suffix), nil
}

// Value returns the value of the current key/value pair, or ErrNotFound if done.
// The caller should not modify the contents of the returned slice, and its
// contents may change on the next call to Next.
func (c *LevelDBCursor) Value() ([]byte, error) {
	if c.isClosed {
		return nil, errors.New("cannot get the value of a closed cursor")
	}
	value := c.ldbIterator.Value()
	if value == nil {
		return nil, errors.Wrapf(database.ErrNotFound, "cannot get the "+
			"value of an exhausted cursor")
	}
	return value, nil
}

// Close releases associated resources.
func (c *LevelDBCursor) Close() error {
	if c.isClosed {
		return errors.New("cannot close an already closed cursor")
	}
	c.isClosed = true
	c.ldbIterator.Release()
	c.ldbIterator = nil
	c.bucket = nil
	return nil
}
