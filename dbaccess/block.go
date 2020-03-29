package dbaccess

import (
<<<<<<< HEAD
	"encoding/hex"
	"github.com/kaspanet/kaspad/database2"
=======
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
>>>>>>> origin/nod-805-database-redesign
	"github.com/pkg/errors"
)

const (
	blockStoreName = "blocks"
)

var (
	blockLocationsBucket = database2.MakeBucket([]byte("block-locations"))
)

// StoreBlock stores the given block in the database.
<<<<<<< HEAD
func StoreBlock(context Context, hash []byte, blockBytes []byte) error {
=======
func StoreBlock(context Context, block *util.Block) error {
>>>>>>> origin/nod-805-database-redesign
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	// Make sure that the block does not already exist.
<<<<<<< HEAD
=======
	hash := block.Hash()
>>>>>>> origin/nod-805-database-redesign
	exists, err := HasBlock(context, hash)
	if err != nil {
		return err
	}
	if exists {
<<<<<<< HEAD
		return errors.Errorf("block %s already exists", hex.EncodeToString(hash))
	}

	// Write the block's bytes to the block store
	blockLocation, err := accessor.AppendToStore(blockStoreName, blockBytes)
=======
		return errors.Errorf("block %s already exists", hash)
	}

	// Write the block's bytes to the block store
	bytes, err := block.Bytes()
	if err != nil {
		return err
	}
	blockLocation, err := accessor.AppendToStore(blockStoreName, bytes)
>>>>>>> origin/nod-805-database-redesign
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
<<<<<<< HEAD
func HasBlock(context Context, hash []byte) (bool, error) {
=======
func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
>>>>>>> origin/nod-805-database-redesign
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
<<<<<<< HEAD
func FetchBlock(context Context, hash []byte) (block []byte, found bool, err error) {
=======
func FetchBlock(context Context, hash *daghash.Hash) (block *util.Block, found bool, err error) {
>>>>>>> origin/nod-805-database-redesign
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

<<<<<<< HEAD
	return bytes, true, nil
}

func blockLocationKey(hash []byte) []byte {
	return blockLocationsBucket.Key(hash)
=======
	block, err = util.NewBlockFromBytes(bytes)
	if err != nil {
		return nil, false, err
	}
	return block, true, nil
}

func blockLocationKey(hash *daghash.Hash) []byte {
	return blockLocationsBucket.Key(hash[:])
>>>>>>> origin/nod-805-database-redesign
}
