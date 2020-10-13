package model

// BlockMessageStore represents a store of MsgBlock
type BlockMessageStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, msgBlock *DomainBlock)
	Block(dbContext DBContextProxy, blockHash *DomainHash) *DomainBlock
	Blocks(dbContext DBContextProxy, blockHashes []*DomainHash) []*DomainBlock
	Delete(dbTx DBTxProxy, blockHash *DomainHash)
}
