package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/state/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGManager interface {
	GHOSTDAG(blockHash *daghash.Hash)
	BlockData(blockHash *daghash.Hash) *ghostdagdatastore.BlockGHOSTDAGData
}
