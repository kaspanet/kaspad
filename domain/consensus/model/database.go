package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DBContextProxy defines a proxy over domain data access
type DBContextProxy interface {
	FetchBlockRelation(blockHash *externalapi.DomainHash) (*BlockRelations, error)
}

// DBTxProxy is a proxy over domain data
// access that requires an open database transaction
type DBTxProxy interface {
	StoreBlockRelation(blockHash *externalapi.DomainHash, blockRelationData *BlockRelations) error
}
