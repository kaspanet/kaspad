package externalapi

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

type BlockWithMetaData struct {
	Block        *DomainBlock
	DAAScore     uint64
	GHOSTDAGData *model.BlockGHOSTDAGData
}
