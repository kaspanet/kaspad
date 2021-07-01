package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type DAAWindowStore interface {
	Store
	IsStaged(stagingArea *StagingArea) bool
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, index uint64, ghostdagData *externalapi.BlockGHOSTDAGDataHashPair)
	DAAWindowBlock(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash, index uint64) (*externalapi.BlockGHOSTDAGDataHashPair, error)
}
