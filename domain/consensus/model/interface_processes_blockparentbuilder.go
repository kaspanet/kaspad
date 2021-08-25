package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockParentBuilder exposes a method to build super-block parents for
// a given set of direct parents
type BlockParentBuilder interface {
	BuildParents(stagingArea *StagingArea, directParents []*externalapi.DomainHash) ([]externalapi.BlockLevelParents, error)
}
