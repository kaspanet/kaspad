package blockparentbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockParentBuilder struct {
	databaseContext model.DBManager
}

// New creates a new instance of a BlockParentBuilder
func New(
	databaseContext model.DBManager,
) model.BlockParentBuilder {

	return &blockParentBuilder{
		databaseContext: databaseContext,
	}
}

func (b blockParentBuilder) BuildParents(stagingArea *model.StagingArea,
	directParents []*externalapi.DomainHash) ([]externalapi.BlockLevelParents, error) {

	return []externalapi.BlockLevelParents{directParents}, nil
}
