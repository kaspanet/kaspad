package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
)

var reachabilityDataBucket = database2.MakeBucket([]byte("reachability"))

// ReachabilityDataCursor opens a cursor over all the
// reachability data entries.
func ReachabilityDataCursor(context Context) (database2.Cursor, error) {
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

	key := reachabilityDataBucket.Key(blockHash[:])
	return accessor.Put(key, reachabilityData)
}

// ClearAllReachabilityData clears all the reachability data
// from database.
func ClearAllReachabilityData(context Context) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	cursor, err := accessor.Cursor(reachabilityDataBucket.Path())
	if err != nil {
		return err
	}

	for ok := cursor.First(); ok; ok = cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		err = accessor.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}
