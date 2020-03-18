package dbaccess

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

const blockStoreName = "blocks"

func StoreBlock(context Context, block *util.Block) error {
	return nil
}

func FetchBlock(context Context, hash *daghash.Hash) ([]byte, error) {
	return nil, nil
}

func HasBlock(context Context, hash *daghash.Hash) (bool, error) {
	return false, nil
}
