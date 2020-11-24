package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelations represents a block's parent/child relations
type BlockRelations struct {
	Parents  []*externalapi.DomainHash
	Children []*externalapi.DomainHash
}

// Clone returns a clone of BlockRelations
func (br *BlockRelations) Clone() *BlockRelations {
	if br == nil {
		return nil
	}

	return &BlockRelations{
		Parents:  externalapi.CloneHashes(br.Parents),
		Children: externalapi.CloneHashes(br.Children),
	}
}
