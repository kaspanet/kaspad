package model

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, msgBlock *DomainBlock)
	Get(dbContext DBContextProxy, blockHash *DomainHash) *DomainBlock
}
