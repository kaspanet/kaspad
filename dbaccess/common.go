package dbaccess

import "github.com/kaspanet/kaspad/database"

func clearBucket(dbTx *TxContext, bucket *database.Bucket) error {
	accessor, err := dbTx.accessor()
	if err != nil {
		return err
	}

	// Collect all of the keys before deleting them. We do this
	// as to not modify the cursor while we're still iterating
	// over it.
	keys := make([][]byte, 0)
	cursor, err := accessor.Cursor(bucket.Path())
	if err != nil {
		return err
	}
	defer cursor.Close()

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

	return nil
}
