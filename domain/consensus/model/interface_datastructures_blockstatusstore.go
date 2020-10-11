package model

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, blockStatus BlockStatus)
	Get(dbContext DBContextProxy, blockHash *DomainHash) BlockStatus
}
