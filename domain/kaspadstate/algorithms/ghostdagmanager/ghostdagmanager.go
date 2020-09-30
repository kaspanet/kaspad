package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGManager interface {
	GHOSTDAG(blockHash *daghash.Hash)
	BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData
}
