package model

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, blockRelationData *BlockRelations) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (*BlockRelations, error)
}
