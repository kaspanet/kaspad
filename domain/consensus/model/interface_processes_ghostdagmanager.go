package model

import "github.com/kaspanet/kaspad/util/daghash"

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager interface {
	GHOSTDAG(blockParents []*daghash.Hash) *BlockGHOSTDAGData
	BlockData(blockHash *daghash.Hash) *BlockGHOSTDAGData
}
