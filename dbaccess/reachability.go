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

	key := reachabilityKey(blockHash)
	return accessor.Put(key, reachabilityData)
}

// ClearReachabilityData clears the reachability data
// from database.
func ClearReachabilityData(dbTx *TxContext) error {
	return clearBucket(dbTx, reachabilityDataBucket)
}

func reachabilityKey(hash *daghash.Hash) []byte {
	return reachabilityDataBucket.Key(hash[:])
}
