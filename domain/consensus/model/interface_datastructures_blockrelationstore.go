package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Store
	StageBlockRelation(blockHash *externalapi.DomainHash, parentHashes []*externalapi.DomainHash)
	StageTips(tipHashess []*externalapi.DomainHash)
	IsAnythingStaged() bool
	BlockRelation(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*BlockRelations, error)
	Tips(dbContext DBContextProxy) ([]*externalapi.DomainHash, error)
}
