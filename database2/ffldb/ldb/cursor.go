package ldb

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBCursor is a thin wrapper around native leveldb iterators.
type LevelDBCursor struct {
	ldbIterator iterator.Iterator
	prefix      []byte

	isClosed bool
}

// Cursor begins a new cursor over the given prefix.
func (db *LevelDB) Cursor(prefix []byte) *LevelDBCursor {
	ldbIterator := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	return &LevelDBCursor{
		ldbIterator: ldbIterator,
		prefix:      prefix,
		isClosed:    false,
	}
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted. Returns false if the cursor is closed.
func (c *LevelDBCursor) Next() bool {
	if c.isClosed {
		return false
	}
	return c.ldbIterator.Next()
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error.
func (c *LevelDBCursor) Error() error {
	return errors.WithStack(c.ldbIterator.Error())
}

// First moves the iterator to the first key/value pair. It returns false if
// such a pair does not exist or if the cursor is closed.
func (c *LevelDBCursor) First() bool {
	if c.isClosed {
		return false
	}
	return c.ldbIterator.First()
}

// Seek moves the iterator to the first key/value pair whose key is greater
// than or equal to the given key. It returns whether such pair exist.
func (c *LevelDBCursor) Seek(key []byte) (bool, error) {
	if c.isClosed {
		return false, errors.New("cannot seek a closed cursor")
	}
	return c.ldbIterator.Seek(key), nil
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (c *LevelDBCursor) Key() ([]byte, error) {
	if c.isClosed {
		return nil, errors.New("cannot get the key of a closed cursor")
	}
	fullKeyPath := c.ldbIterator.Key()
	key := bytes.TrimPrefix(fullKeyPath, c.prefix)
	return key, nil
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (c *LevelDBCursor) Value() ([]byte, error) {
	if c.isClosed {
		return nil, errors.New("cannot get the value of a closed cursor")
	}
	return c.ldbIterator.Value(), nil
}

// Close releases associated resources.
func (c *LevelDBCursor) Close() error {
	if c.isClosed {
		return errors.New("cannot close an already closed cursor")
	}
	c.isClosed = true

	c.ldbIterator.Release()
	return nil
}
