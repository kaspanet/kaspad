package dbaccess

import (
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

	blockIndexKey := blockIndexKey(blockHash)
	return db.Put(blockIndexKey, dbBlock)
}

func BlockIndexCursor(context Context) (database2.Cursor, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	return db.Cursor(blockIndexBucket.Path())
}

func blockIndexKey(hash *daghash.Hash) []byte {
	return blockIndexBucket.Key(hash[:])
}
