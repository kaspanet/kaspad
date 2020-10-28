package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Store
	StageBlockRelation(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash)
	StageTips(tipHashess []*externalapi.DomainHash)
	IsAnythingStaged() bool
	BlockRelation(dbContext DBReader, blockHash *externalapi.DomainHash) (*BlockRelations, error)
	Tips(dbContext DBReader) ([]*externalapi.DomainHash, error)
}
