package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

var utxoDiffsBucket = database.MakeBucket([]byte("utxo-diffs"))

func utxoDiffKey(hash *daghash.Hash) []byte {
	return utxoDiffsBucket.Key(hash[:])
}

// StoreUTXODiffData stores the UTXO diff data of a block by its hash.
func StoreUTXODiffData(context Context, blockHash *daghash.Hash, diffData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := utxoDiffKey(blockHash)
	return accessor.Put(key, diffData)
}

// RemoveDiffData removes the UTXO diff data from the block with the
// given hash.
func RemoveDiffData(context Context, blockHash *daghash.Hash) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := utxoDiffKey(blockHash)
	return accessor.Delete(key)
}

// FetchUTXODiffData returns the UTXO diff data of a block by its hash.
func FetchUTXODiffData(context Context, blockHash *daghash.Hash) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	key := utxoDiffKey(blockHash)
	diffData, err := accessor.Get(key)
	if IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "couldn't find diff data for block %s", blockHash)
	}
	return diffData, err
}
