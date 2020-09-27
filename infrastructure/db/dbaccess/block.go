package dbaccess

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

var (
	blocksBucket = database.MakeBucket([]byte("blocks"))
)

func blockKey(hash *daghash.Hash) *database.Key {
	return blocksBucket.Key(hash[:])
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
	err = accessor.Put(blockKey(hash), blockBytes)
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

	return accessor.Has(blockKey(hash))
}

// FetchBlock returns the block of the given hash. Returns
// ErrNotFound if the block had not been previously inserted
// into the database.
func FetchBlock(context Context, hash *daghash.Hash) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Get(blockKey(hash))
}
