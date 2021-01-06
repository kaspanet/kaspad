package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelations represents a block's parent/child relations
type BlockRelations struct {
	Parents  []*externalapi.DomainHash
	Children []*externalapi.DomainHash
}

// Clone returns a clone of BlockRelations
func (br *BlockRelations) Clone() *BlockRelations {
	return &BlockRelations{
		Parents:  externalapi.CloneHashes(br.Parents),
		Children: externalapi.CloneHashes(br.Children),
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = &BlockRelations{[]*externalapi.DomainHash{}, []*externalapi.DomainHash{}}

// Equal returns whether br equals to other
func (br *BlockRelations) Equal(other *BlockRelations) bool {
	if br == nil || other == nil {
		return br == other
	}

	if !externalapi.HashesEqual(br.Parents, other.Parents) {
		return false
	}

	if !externalapi.HashesEqual(br.Children, other.Children) {
		return false
	}

	return true
}
