package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockHeaderStore represents a store of block headers
type BlockHeaderStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, blockHeader *externalapi.DomainBlockHeader)
	IsStaged() bool
	BlockHeader(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error)
	HasBlockHeader(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (bool, error)
	BlockHeaders(dbContext DBContextProxy, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlockHeader, error)
	Delete(dbTx DBTxProxy, blockHash *externalapi.DomainHash) error
}
