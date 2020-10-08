package model

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, blockRelationData *BlockRelations)
	Get(dbContext ContextProxy, blockHash *DomainHash) *BlockRelations
}
