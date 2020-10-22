package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Stage(blockHash *externalapi.DomainHash, blockStatus BlockStatus)
	IsStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (BlockStatus, error)
	Exists(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (bool, error)
}
