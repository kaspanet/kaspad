package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
)

var (
	blockIndexBucket = database2.MakeBucket([]byte("block-index"))
)

// StoreIndexBlock stores a block in block-index
// representation to the database.
func StoreIndexBlock(context Context, blockIndexKey []byte, block []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	return accessor.Put(blockIndexKey, block)
}

// BlockIndexCursor opens a cursor over all the blocks-index
// blocks that have been previously added to the database.
func BlockIndexCursor(context Context) (database2.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(blockIndexBucket.Path())
}

// BlockIndexCursor opens a cursor over blocks-index blocks
// starting from the block with the given blockHash and
// blockBlueScore.
func BlockIndexCursorFrom(context Context, blockIndexKey []byte) (cursor database2.Cursor, found bool, err error) {
	cursor, err = BlockIndexCursor(context)
	if err != nil {
		return nil, false, err
	}

	found, err = cursor.Seek(blockIndexKey)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	return cursor, true, nil
}
