package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
)

var multisetBucket = database2.MakeBucket([]byte("multiset"))

// MultisetCursor opens a cursor over all the
// multiset entries.
func MultisetCursor(context Context) (database2.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(multisetBucket.Path())
}

// StoreMultiset stores the multiset of a block by its hash.
func StoreMultiset(context Context, blockHash *daghash.Hash, multiset []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := multisetBucket.Key(blockHash[:])
	return accessor.Put(key, multiset)
}

// MultisetExists returns whether the multiset of
// the given block exists in the database.
func MultisetExists(context Context, blockHash *daghash.Hash) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := multisetBucket.Key(blockHash[:])
	_, err = accessor.Get(key)
	if IsNotFoundError(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
