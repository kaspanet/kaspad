package model

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, blockStatus BlockStatus)
	Get(dbContext ContextProxy, blockHash *DomainHash) BlockStatus
}
