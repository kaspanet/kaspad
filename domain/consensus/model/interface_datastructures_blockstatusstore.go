package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockStatusStore represents a store of BlockStatuses
type BlockStatusStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, blockStatus externalapi.BlockStatus)
	IsStaged() bool
	Get(dbContext DBReader, blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error)
	Exists(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
}
