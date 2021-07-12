package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, blockGHOSTDAGData *externalapi.BlockGHOSTDAGData, isMetaData bool)
	IsStaged(stagingArea *StagingArea) bool
	Get(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash, isMetaData bool) (*externalapi.BlockGHOSTDAGData, error)
}
