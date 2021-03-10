package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeaderSelectedTipStore represents a store of the headers selected tip
type HeaderSelectedTipStore interface {
	Store
	Stage(selectedTip *externalapi.DomainHash)
	IsStaged() bool
	HeadersSelectedTip(dbContext DBReader) (*externalapi.DomainHash, error)
	Has(dbContext DBReader) (bool, error)
}
