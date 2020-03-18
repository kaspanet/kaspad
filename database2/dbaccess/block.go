package dbaccess

import (
	"github.com/kaspanet/kaspad/database2/ffldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

const (
	blockStoreName = "blocks"
)

var (
	blockLocationsBucketName = []byte("block-locations")
)

func StoreBlock(context Context, block *util.Block) error {
	db, err := context.db()
	if err != nil {
		return err
	}
	bytes, err := block.Bytes()
	if err != nil {
		return err
	}

	blockLocation, err := db.AppendFlatData(blockStoreName, bytes)
	if err != nil {
		return err
	}
	blockLocationsKey := blockLocationKey(block.Hash())
	return db.Put(blockLocationsKey, blockLocation)
}

func FetchBlock(context Context, hash *daghash.Hash) ([]byte, error) {
	return nil, nil
}

func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
	return false, nil
}

func blockLocationKey(hash *daghash.Hash) []byte {
	return ffldb.Key(hash[:], blockLocationsBucketName)
}
