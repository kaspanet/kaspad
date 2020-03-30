package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

const (
	blockStoreName = "blocks"
)

var (
	blockLocationsBucket = database2.MakeBucket([]byte("block-locations"))
)

// StoreBlock stores the given block in the database.
func StoreBlock(context Context, block *util.Block) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	// Make sure that the block does not already exist.
	hash := block.Hash()
	exists, err := HasBlock(context, hash)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("block %s already exists", hash)
	}

	// Write the block's bytes to the block store
	bytes, err := block.Bytes()
	if err != nil {
		return err
	}
	blockLocation, err := accessor.AppendToStore(blockStoreName, bytes)
	if err != nil {
		return err
	}

	// Write the block's hash to the blockLocations bucket
	blockLocationsKey := blockLocationKey(hash)
	err = accessor.Put(blockLocationsKey, blockLocation)
	if err != nil {
		return err
	}

	return nil
}

// HasBlock returns whether the block of the given hash has been
// previously inserted into the database.
func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	blockLocationsKey := blockLocationKey(hash)

	return accessor.Has(blockLocationsKey)
}

// FetchBlock returns the block of the given hash. Returns
// ErrNotFound if the block had not been previously inserted
// into the database.
func FetchBlock(context Context, hash *daghash.Hash) (*util.Block, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	blockLocationsKey := blockLocationKey(hash)
	blockLocation, err := accessor.Get(blockLocationsKey)
	if err != nil {
		return nil, err
	}
	bytes, err := accessor.RetrieveFromStore(blockStoreName, blockLocation)
	if err != nil {
		return nil, err
	}

	block, err := util.NewBlockFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func blockLocationKey(hash *daghash.Hash) []byte {
	return blockLocationsBucket.Key(hash[:])
}
