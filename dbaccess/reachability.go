package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
)

var reachabilityDataBucket = database.MakeBucket([]byte("reachability"))

// ReachabilityDataCursor opens a cursor over all the
// reachability data entries.
func ReachabilityDataCursor(context Context) (database.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(reachabilityDataBucket.Path())
}

// StoreReachabilityData stores the reachability data of a block by its hash.
func StoreReachabilityData(context Context, blockHash *daghash.Hash, reachabilityData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := reachabilityKey(blockHash)
	return accessor.Put(key, reachabilityData)
}

// ClearReachabilityData clears the reachability data
// from database.
func ClearReachabilityData() error {
	dbTx, err := NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	accessor, err := dbTx.accessor()
	if err != nil {
		return err
	}

	// Collect all of the keys before deleting them. We do this
	// as to not modify the cursor while we're still iterating
	// over it.
	keys := make([][]byte, 0)
	cursor, err := accessor.Cursor(reachabilityDataBucket.Path())
	if err != nil {
		return err
	}
	for cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}
		keys = append(keys, key)
	}

	// Delete all of the keys
	for _, key := range keys {
		err := accessor.Delete(key)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

func reachabilityKey(hash *daghash.Hash) []byte {
	return reachabilityDataBucket.Key(hash[:])
}
