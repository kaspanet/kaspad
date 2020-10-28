package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Store
	StageBlockRelation(blockHash *externalapi.DomainHash, blockRelations *BlockRelations)
	IsStaged() bool
	BlockRelation(dbContext DBReader, blockHash *externalapi.DomainHash) (*BlockRelations, error)
}
