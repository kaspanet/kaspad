package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

var (
	acceptanceIndexBucket = database.MakeBucket([]byte("acceptance-index"))
)

// StoreAcceptanceData stores the given acceptanceData in the database.
func StoreAcceptanceData(context Context, hash *daghash.Hash, acceptanceData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := acceptanceIndexKey(hash)
	return accessor.Put(key, acceptanceData)
}

// HasAcceptanceData returns whether the acceptanceData of the given hash
// has been previously inserted into the database.
func HasAcceptanceData(context Context, hash *daghash.Hash) (bool, error) {
	accessor, err := context.accessor()
	if err != nil {
		return false, err
	}

	key := acceptanceIndexKey(hash)
	return accessor.Has(key)
}

// FetchAcceptanceData returns the acceptanceData of the given hash.
// Returns ErrNotFound if the acceptanceData had not been previously
// inserted into the database.
func FetchAcceptanceData(context Context, hash *daghash.Hash) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	key := acceptanceIndexKey(hash)
	acceptanceData, err := accessor.Get(key)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "acceptance data not found for hash %s", hash)
		}
		return nil, err
	}

	return acceptanceData, nil
}

func acceptanceIndexKey(hash *daghash.Hash) []byte {
	return acceptanceIndexBucket.Key(hash[:])
}

// DropAcceptanceIndex completely removes all acceptanceData entries.
func DropAcceptanceIndex() error {
	context, err := NewTx()
	if err != nil {
		return err
	}
	defer context.RollbackUnlessClosed()
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	// Collect all of the keys before deleting them. We do this
	// as to not modify the cursor while we're still iterating
	// over it.
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
