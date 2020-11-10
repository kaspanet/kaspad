package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Store
	StageBlockRelation(blockHash *externalapi.DomainHash, blockRelations *BlockRelations) error
	IsStaged() bool
	BlockRelation(dbContext DBReader, blockHash *externalapi.DomainHash) (*BlockRelations, error)
	Has(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
}
