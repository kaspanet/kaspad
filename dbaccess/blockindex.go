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
	db, err := context.db()
	if err != nil {
		return err
	}

	blockIndexKey := blockIndexKey(blockHash, blockBlueScore)
	return db.Put(blockIndexKey, block)
}

// BlockIndexCursor opens a cursor over all the blocks-index
// blocks that have been previously added to the database.
func BlockIndexCursor(context Context) (database2.Cursor, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	return db.Cursor(blockIndexBucket.Path())
}

func blockIndexKey(blockHash []byte, blueScore uint64) []byte {
	key := make([]byte, 40)
	binary.BigEndian.PutUint64(key[0:8], blueScore)
	copy(key[8:40], blockHash[:])

	return blockIndexBucket.Key(key)
}
