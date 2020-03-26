package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

var feeBucket = database2.MakeBucket([]byte("fees"))

// FetchFeeData returns the fee data of a block by its hash.
func FetchFeeData(context Context, blockHash *daghash.Hash) ([]byte, error) {
	db, err := context.db()
	if err != nil {
		return nil, err
	}

	key := feeBucket.Key(blockHash[:])
	feeData, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	if feeData == nil {
		return nil, errors.Errorf("No fee data found for block %s", blockHash)
	}

	return feeData, nil
}

// StoreFeeData stores the fee data of a block by its hash.
func StoreFeeData(context Context, blockHash *daghash.Hash, feeData []byte) error {
	db, err := context.db()
	if err != nil {
		return err
	}

	key := feeBucket.Key(blockHash[:])
	return db.Put(key, feeData)
}
