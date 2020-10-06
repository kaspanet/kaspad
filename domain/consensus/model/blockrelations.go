package model

import "github.com/kaspanet/kaspad/util/daghash"

// BlockRelations represents a block's parent/child relations
type BlockRelations struct {
	Parents  []*daghash.Hash
	Children []*daghash.Hash
}
