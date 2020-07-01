package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
)

var reachabilityDataBucket = database.MakeBucket([]byte("reachability"))
var reachabilityReindexKey = database.MakeBucket().Key([]byte("reachability-reindex-root"))

func reachabilityKey(hash *daghash.Hash) *database.Key {
	return reachabilityDataBucket.Key(hash[:])
}

// ReachabilityDataCursor opens a cursor over all the
// reachability data entries.
func ReachabilityDataCursor(context Context) (database.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(reachabilityDataBucket)
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

// StoreReachabilityReindexRoot stores the reachability reindex root in the database.
func StoreReachabilityReindexRoot(context Context, reachabilityReindexRoot *daghash.Hash) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}
	return accessor.Put(reachabilityReindexKey, reachabilityReindexRoot[:])
}

// FetchReachabilityReindexRoot retrieves the reachability reindex root from the database.
// Returns ErrNotFound if the state is missing from the database.
func FetchReachabilityReindexRoot(context Context) (*daghash.Hash, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}
	bytes, err := accessor.Get(reachabilityReindexKey)
	if err != nil {
		return nil, err
	}
	return daghash.NewHash(bytes)
}
