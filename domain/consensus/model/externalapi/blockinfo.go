package externalapi

import "github.com/kaspanet/kaspad/domain/consensus/model"

type BlockInfo struct {
	Exists      bool
	BlockStatus *model.BlockStatus

	IsHeaderInPruningPointFutureAndVirtualPast bool
}
