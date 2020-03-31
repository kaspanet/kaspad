package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
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
	acceptanceData, err := accessor.Get(key)
	if err != nil {
		if database2.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "acceptance data not found for hash %s", hash)
		}
		return nil, err
	}

	return acceptanceData, nil
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
