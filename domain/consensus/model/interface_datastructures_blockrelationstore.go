package model

// BlockRelationStore represents a store of BlockRelations
type BlockRelationStore interface {
	Update(dbTx DBTxProxy, blockHash *DomainHash, parentHashes []*DomainHash) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (*BlockRelations, error)
}
