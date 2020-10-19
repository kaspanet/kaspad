package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockStore represents a store of blocks
type BlockStore interface {
	Insert(dbTx DBTxProxy, blockHash *externalapi.DomainHash, block *externalapi.DomainBlock) error
	Block(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error)
	Blocks(dbContext DBContextProxy, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error)
	Delete(dbTx DBTxProxy, blockHash *externalapi.DomainHash) error
}
