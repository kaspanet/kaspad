package model

// DBContextProxy defines a proxy over domain data access
type DBContextProxy interface {
	FetchBlockRelation(blockHash *DomainHash) (*BlockRelations, error)
}

// DBTxProxy is a proxy over domain data
// access that requires an open database transaction
type DBTxProxy interface {
	StoreBlockRelation(blockHash *DomainHash, blockRelationData *BlockRelations) error
}
