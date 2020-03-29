package dbaccess

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/database2"
	"github.com/pkg/errors"
)

const (
	blockStoreName = "blocks"
)

var (
	blockLocationsBucket = database2.MakeBucket([]byte("block-locations"))
)

// StoreBlock stores the given block in the database.
func StoreBlock(context Context, hash []byte, blockBytes []byte) error {
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
		return errors.Errorf("block %s already exists", hex.EncodeToString(hash))
	}

	// Write the block's bytes to the block store
	blockLocation, err := db.AppendToStore(blockStoreName, blockBytes)
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
func HasBlock(context Context, hash []byte) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	blockLocationsKey := blockLocationKey(hash)

	return accessor.Has(blockLocationsKey)
}

// FetchBlock returns the block of the given hash. Returns
// found=false if the block had not been previously inserted
// into the database.
func FetchBlock(context Context, hash []byte) (block []byte, found bool, err error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, false, err
	}

	blockLocationsKey := blockLocationKey(hash)
	blockLocation, found, err := accessor.Get(blockLocationsKey)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	bytes, found, err := accessor.RetrieveFromStore(blockStoreName, blockLocation)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	return bytes, nil
}

func blockLocationKey(hash []byte) []byte {
	return blockLocationsBucket.Key(hash)
}
