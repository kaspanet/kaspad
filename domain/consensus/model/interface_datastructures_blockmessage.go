package model

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, msgBlock *DomainBlock)
	Get(dbContext ContextProxy, blockHash *DomainHash) *DomainBlock
}
