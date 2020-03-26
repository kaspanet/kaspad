package dbaccess

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/database2"
)

var (
	blockIndexBucket = database2.MakeBucket([]byte("block-index"))
)

func StoreIndexBlock(context Context, blockHash []byte, blockBlueScore uint64, block []byte) error {
	db, err := context.db()
	if err != nil {
		return err
	}

	blockIndexKey := blockIndexKey(blockHash, blockBlueScore)
	return db.Put(blockIndexKey, block)
}

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
