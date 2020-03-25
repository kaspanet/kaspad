package dbaccess

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/dbaccess/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

var (
	blockIndexBucket = database2.MakeBucket([]byte("block-index"))
)

func StoreIndexBlock(context Context, blockHash *daghash.Hash, block *model.BlockNode) error {
	db, err := context.db()
	if err != nil {
		return err
	}

	dbBlock, err := model.SerializeBlockNode(block)
	if err != nil {
		return err
	}

	blockIndexKey := blockIndexKey(blockHash, block.BlueScore)
	return db.Put(blockIndexKey, dbBlock)
}

func BlockIndexCursor(context Context) (database2.Cursor, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	return db.Cursor(blockIndexBucket.Path())
}

func blockIndexKey(blockHash *daghash.Hash, blueScore uint64) []byte {
	key := make([]byte, daghash.HashSize+8)
	binary.BigEndian.PutUint64(key[0:8], blueScore)
	copy(key[8:daghash.HashSize+8], blockHash[:])

	return blockIndexBucket.Key(key)
}
