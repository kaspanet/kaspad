package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, blockGHOSTDAGData *BlockGHOSTDAGData)
	IsStaged() bool
	Get(dbContext DBReader, blockHash *externalapi.DomainHash) (*BlockGHOSTDAGData, error)
}
