package dbaccess

import (
	"github.com/kaspanet/kaspad/database2/ffldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

const (
	blockStoreName = "blocks"
)

var (
	blockLocationsBucketName    = []byte("block-locations")
	currentBlockLocationKeyName = []byte("current-block-location")
)

func StoreBlock(context Context, block *util.Block) error {
	db, err := context.db()
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

	// Save a reference to the current block location in case
	// we fail and need to rollback.
	previousBlockLocation := db.CurrentFlatDataLocation(blockStoreName)
	rollback := func() error {
		return db.RollbackFlatData(blockStoreName, previousBlockLocation)
	}

	// Write the block's bytes to the block store and rollback
	// if there's an error.
	bytes, err := block.Bytes()
	if err != nil {
		return err
	}
	blockLocation, err := db.AppendFlatData(blockStoreName, bytes)
	if err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	// Write the block's hash to the blockLocations bucket and
	// rollback if there's an error.
	blockLocationsKey := blockLocationKey(hash)
	err = db.Put(blockLocationsKey, blockLocation)
	if err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	// Write the new block location. We use it to reconcile the
	// block store and the block locations bucket when kaspad
	// restarts. Rollback if this fails.
	currentBlockLocation := db.CurrentFlatDataLocation(blockStoreName)
	err = db.Put(currentBlockLocationKeyName, currentBlockLocation)
	if err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	return nil
}

func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
	db, err := context.db()
	if err != nil {
		return false, err
	}

	blockLocationsKey := blockLocationKey(hash)

	return db.Has(blockLocationsKey)
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

func blockLocationKey(hash *daghash.Hash) []byte {
	return ffldb.Key(hash[:], blockLocationsBucketName)
}
