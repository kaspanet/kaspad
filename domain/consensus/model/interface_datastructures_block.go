package model

// BlockStore represents a store of blocks
type BlockStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, block *DomainBlock) error
	Block(dbContext DBContextProxy, blockHash *DomainHash) (*DomainBlock, error)
	Blocks(dbContext DBContextProxy, blockHashes []*DomainHash) ([]*DomainBlock, error)
	Delete(dbTx DBTxProxy, blockHash *DomainHash) error
}
