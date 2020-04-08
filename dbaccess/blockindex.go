package dbaccess

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/database"
	"github.com/pkg/errors"
)

var (
	blockIndexBucket = database.MakeBucket([]byte("block-index"))
)

// StoreIndexBlock stores a block in block-index
// representation to the database.
func StoreIndexBlock(context Context, blockIndexKey []byte, block []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := blockIndexBucket.Key(blockIndexKey)
	return accessor.Put(key, block)
}

// BlockIndexCursor opens a cursor over all the blocks-index
// blocks that have been previously added to the database.
func BlockIndexCursor(context Context) (database.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(blockIndexBucket)
}

// BlockIndexCursorFrom opens a cursor over blocks-index blocks
// starting from the block with the given blockHash and
// blockBlueScore. Returns ErrNotFound if blockIndexKey is missing
// from the database.
func BlockIndexCursorFrom(context Context, blockIndexKey []byte) (database.Cursor, error) {
	cursor, err := BlockIndexCursor(context)
	if err != nil {
		return nil, err
	}

	key := blockIndexBucket.Key(blockIndexKey)
	err = cursor.Seek(key)
	if IsNotFoundError(err) {
		cursor.Close()
		return nil, errors.Wrapf(database.ErrNotFound,
			"entry not found for %s", hex.EncodeToString(blockIndexKey))
	}
	if err != nil {
		return nil, err
	}

	return cursor, nil
}
