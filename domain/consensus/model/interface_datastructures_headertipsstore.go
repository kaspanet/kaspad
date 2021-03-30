package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeaderSelectedTipStore represents a store of the headers selected tip
type HeaderSelectedTipStore interface {
	Store
	Stage(stagingArea *StagingArea, selectedTip *externalapi.DomainHash)
	IsStaged(stagingArea *StagingArea) bool
	HeadersSelectedTip(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	Has(dbContext DBReader, stagingArea *StagingArea) (bool, error)
}
