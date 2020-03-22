package ffldb

import (
	"github.com/kaspanet/kaspad/database2/ffldb/ldb"
)

// cursor is an ffldb cursor.
type cursor struct {
	ldbCursor *ldb.LevelDBCursor
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted. Returns false if the cursor is closed.
func (c *cursor) Next() bool {
	return c.ldbCursor.Next()
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error.
func (c *cursor) Error() error {
	return c.ldbCursor.Error()
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (c *cursor) Key() ([]byte, error) {
	return c.ldbCursor.Key()
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (c *cursor) Value() ([]byte, error) {
	return c.ldbCursor.Value()
}

// Close releases associated resources.
func (c *cursor) Close() error {
	return c.ldbCursor.Close()
}
