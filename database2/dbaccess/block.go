package dbaccess

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

func StoreBlock(context Context, block *util.Block) error {
	return nil
}

func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
	return false, nil
}

func FetchBlock(context Context, hash *daghash.Hash) ([]byte, error) {
	return nil, nil
}

func FetchBlocks(context Context, hashes []*daghash.Hash) ([][]byte, error) {
	return nil, nil
}
