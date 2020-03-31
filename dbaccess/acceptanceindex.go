package dbaccess

import "github.com/kaspanet/kaspad/database2"

var (
	acceptanceIndexBucket = database2.MakeBucket([]byte("acceptance-index"))
)

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
