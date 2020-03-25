package ldb

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBCursor is a thin wrapper around native leveldb iterators.
type LevelDBCursor struct {
	ldbIterator iterator.Iterator

	isClosed bool
}

// Cursor begins a new cursor over the given prefix.
func (db *LevelDB) Cursor(prefix []byte) *LevelDBCursor {
	ldbIterator := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	return &LevelDBCursor{
		ldbIterator: ldbIterator,
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

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (c *LevelDBCursor) Key() ([]byte, error) {
	if c.isClosed {
		return nil, errors.New("cannot get the key of a closed cursor")
	}
	return c.ldbIterator.Key(), nil
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
