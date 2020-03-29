package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
)

// reachabilityDataBucketName is the name of the database bucket used to house the
// reachability tree nodes and future covering sets of blocks.
var reachabilityDataBucket = database2.MakeBucket([]byte("reachability"))

// ReachabilityDataCursor opens a cursor over all the
// reachability data entries.
func ReachabilityDataCursor(context Context) (database2.Cursor, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	return db.Cursor(reachabilityDataBucket.Path())
}

// StoreReachabilityData stores the reachability data of a block by its hash.
func StoreReachabilityData(context Context, blockHash *daghash.Hash, data []byte) error {
	db, err := context.db()
	if err != nil {
		return err
	}

	key := reachabilityDataBucket.Key(blockHash[:])
	return db.Put(key, data)
}
