package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockStore represents a store of blocks
type BlockStore interface {
	Stage(blockHash *externalapi.DomainHash, block *externalapi.DomainBlock)
	IsStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	Block(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error)
	HasBlock(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (bool, error)
	Blocks(dbContext DBContextProxy, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error)
	Delete(dbTx DBTxProxy, blockHash *externalapi.DomainHash) error
}
