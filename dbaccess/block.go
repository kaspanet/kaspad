package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

const (
	blockStoreName = "blocks"
)

var (
	blockLocationsBucket = database.MakeBucket([]byte("block-locations"))
)

func blockLocationKey(hash *daghash.Hash) *database.Key {
	return blockLocationsBucket.Key(hash[:])
}

// StoreBlock stores the given block in the database.
func StoreBlock(context *TxContext, hash *daghash.Hash, blockBytes []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	// Make sure that the block does not already exist.
	exists, err := HasBlock(context, hash)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("block %s already exists", hash)
	}

	// Write the block's bytes to the block store
	blockLocation, err := accessor.AppendToStore(blockStoreName, blockBytes)
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
func FetchBlock(context Context, hash *daghash.Hash) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	blockLocationsKey := blockLocationKey(hash)
	blockLocation, err := accessor.Get(blockLocationsKey)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, errors.Wrapf(err,
				"block %s not found", hash)
		}
		return nil, err
	}
	bytes, err := accessor.RetrieveFromStore(blockStoreName, blockLocation)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
