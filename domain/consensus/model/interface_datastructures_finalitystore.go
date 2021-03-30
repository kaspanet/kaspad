package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// FinalityStore represents a store for finality data
type FinalityStore interface {
	Store
	IsStaged(stagingArea *StagingArea) bool
	StageFinalityPoint(stagingArea *StagingArea, blockHash *externalapi.DomainHash, finalityPointHash *externalapi.DomainHash)
	FinalityPoint(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
}
