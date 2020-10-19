package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelations represents a block's parent/child relations
type BlockRelations struct {
	Parents  []*externalapi.DomainHash
	Children []*externalapi.DomainHash
}
