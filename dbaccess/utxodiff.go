package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
)

var utxoDiffsBucketName = database2.MakeBucket([]byte("utxodiffs"))

// StoreUTXODiff stores the UTXO diff data of a block by its hash.
func StoreUTXODiffData(context Context, blockHash *daghash.Hash, diffData []byte) error {
	db, err := context.db()
	if err != nil {
		return err
	}

	key := utxoDiffsBucketName.Key(blockHash[:])
	return db.Put(key, diffData)
}

// RemoveDiffData removes the UTXO diff data from the block with the
// given hash.
func RemoveDiffData(context Context, blockHash *daghash.Hash) error {
	db, err := context.db()
	if err != nil {
		return err
	}

	key := utxoDiffsBucketName.Key(blockHash[:])
	return db.Delete(key)
}

// FetchUTXODiffData returns the UTXO diff data of a block by its hash.
func FetchUTXODiffData(context Context, blockHash *daghash.Hash) ([]byte, bool, error) {
	db, err := context.db()
	if err != nil {
		return nil, false, err
	}

	key := feeBucket.Key(blockHash[:])
	diffData, err := db.Get(key)
	if err != nil {
		return nil, false, err
	}

	return diffData, nil
}
