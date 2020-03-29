package dbaccess

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/database2"
)

var (
	blockIndexBucket = database2.MakeBucket([]byte("block-index"))
)

// StoreIndexBlock stores a block in block-index
// representation to the database.
func StoreIndexBlock(context Context, blockHash []byte, blockBlueScore uint64, block []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	blockIndexKey := blockIndexKey(blockHash, blockBlueScore)
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

func blockIndexKey(blockHash []byte, blueScore uint64) []byte {
	key := make([]byte, 40)
	binary.BigEndian.PutUint64(key[0:8], blueScore)
	copy(key[8:40], blockHash[:])

	return blockIndexBucket.Key(key)
}
