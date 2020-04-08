package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
)

var multisetBucket = database.MakeBucket([]byte("multiset"))

func multisetKey(hash *daghash.Hash) *database.Key {
	return multisetBucket.Key(hash[:])
}

// MultisetCursor opens a cursor over all the
// multiset entries.
func MultisetCursor(context Context) (database.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(multisetBucket)
}

// StoreMultiset stores the multiset of a block by its hash.
func StoreMultiset(context Context, blockHash *daghash.Hash, multiset []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := multisetKey(blockHash)
	return accessor.Put(key, multiset)
}

// HasMultiset returns whether the multiset of
// the given block exists in the database.
func HasMultiset(context Context, blockHash *daghash.Hash) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := multisetKey(blockHash)
	return accessor.Has(key)
}
