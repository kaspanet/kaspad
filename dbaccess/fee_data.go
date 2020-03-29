package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util/daghash"
)

var feeBucket = database2.MakeBucket([]byte("fees"))

// FetchFeeData returns the fee data of a block by its hash.
func FetchFeeData(context Context, blockHash *daghash.Hash) (feeData []byte, found bool, err error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, false, err
	}

	key := feeBucket.Key(blockHash[:])
	return accessor.Get(key)
}

// StoreFeeData stores the fee data of a block by its hash.
func StoreFeeData(context Context, blockHash *daghash.Hash, feeData []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := feeBucket.Key(blockHash[:])
	return accessor.Put(key, feeData)
}
