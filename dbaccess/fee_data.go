package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

var feeBucket = database.MakeBucket([]byte("fees"))

func feeDataKey(hash *daghash.Hash) *database.Key {
	return feeBucket.Key(hash[:])
}

// FetchFeeData returns the fee data of a block by its hash.
func FetchFeeData(context Context, blockHash *daghash.Hash) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	key := feeDataKey(blockHash)
	feeData, err := accessor.Get(key)
	if IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "couldn't find fee data for block %s", blockHash)
	}
	return feeData, err
}

// StoreFeeData stores the fee data of a block by its hash.
func StoreFeeData(context Context, blockHash *daghash.Hash, feeData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := feeDataKey(blockHash)
	return accessor.Put(key, feeData)
}
