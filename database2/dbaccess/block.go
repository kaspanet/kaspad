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

func FetchBlock(context Context, hash *daghash.Hash) (*util.Block, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	blockLocationsKey := blockLocationKey(hash)
	blockLocation, err := db.Get(blockLocationsKey)
	if err != nil {
		return nil, err
	}
	bytes, err := db.RetrieveFlatData(blockStoreName, blockLocation)
	if err != nil {
		return nil, err
	}

	return util.NewBlockFromBytes(bytes)
}

func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
	db, err := context.db()
	if err != nil {
		return false, err
	}

	blockLocationsKey := blockLocationKey(hash)

	return db.Has(blockLocationsKey)
}

func blockLocationKey(hash *daghash.Hash) []byte {
	return ffldb.Key(hash[:], blockLocationsBucketName)
}
