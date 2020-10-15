package model

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, blockStatus BlockStatus) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (BlockStatus, error)
	Exists(dbContext DBContextProxy, blockHash *DomainHash) (bool, error)
}
