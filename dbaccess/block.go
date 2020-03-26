package dbaccess

import (
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
	db, err := context.db()
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
	blockLocation, err := db.AppendToStore(blockStoreName, blockBytes)
	if err != nil {
		return err
	}

	// Write the block's hash to the blockLocations bucket
	blockLocationsKey := blockLocationKey(hash)
	err = db.Put(blockLocationsKey, blockLocation)
	if err != nil {
		return err
	}

	return nil
}

// HasBlock returns whether the block of the given hash has been
// previously inserted into the database.
func HasBlock(context Context, hash []byte) (bool, error) {
	db, err := context.db()
	if err != nil {
		return false, err
	}

	blockLocationsKey := blockLocationKey(hash)

	return db.Has(blockLocationsKey)
}

// FetchBlock returns the block of the given hash. Returns an
// error if the block had not been previously inserted into the
// database.
func FetchBlock(context Context, hash []byte) ([]byte, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	blockLocationsKey := blockLocationKey(hash)
	blockLocation, err := db.Get(blockLocationsKey)
	if err != nil {
		return nil, err
	}
	if blockLocation == nil {
		return nil, errors.Errorf("block %s not found", hash)
	}
	bytes, err := db.RetrieveFromStore(blockStoreName, blockLocation)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func blockLocationKey(hash []byte) []byte {
	return blockLocationsBucket.Key(hash)
}
