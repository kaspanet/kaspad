package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
)

var (
	acceptanceIndexBucket = database2.MakeBucket([]byte("acceptance-index"))
)

func StoreAcceptanceData(context Context, hash *daghash.Hash, acceptanceData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := acceptanceIndexKey(hash)
	return accessor.Put(key, acceptanceData)
}

func HasAcceptanceData(context Context, hash *daghash.Hash) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := acceptanceIndexKey(hash)
	return accessor.Has(key)
}

func FetchAcceptanceData(context Context, hash *daghash.Hash) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	key := acceptanceIndexKey(hash)
	return accessor.Get(key)
}

func acceptanceIndexKey(hash *daghash.Hash) []byte {
	return acceptanceIndexBucket.Key(hash[:])
}

func ClearAcceptanceIndex() error {
	context, err := NewTx()
	if err != nil {
		return err
	}
	defer context.RollbackUnlessClosed()
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	// Collect all of the keys
	keys := make([][]byte, 0)
	cursor, err := accessor.Cursor(acceptanceIndexBucket.Path())
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

	return context.Commit()
}
