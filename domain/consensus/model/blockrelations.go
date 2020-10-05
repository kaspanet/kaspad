package model

import "github.com/kaspanet/kaspad/util/daghash"

// BlockRelations ...
type BlockRelations struct {
	Parents  []*daghash.Hash
	Children []*daghash.Hash
}
