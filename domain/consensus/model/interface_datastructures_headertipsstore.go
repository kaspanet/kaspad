package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeaderTipsStore represents a store of the header tips
type HeaderTipsStore interface {
	Store
	Stage(tips []*externalapi.DomainHash)
	IsStaged() bool
	Tips(dbContext DBReader) ([]*externalapi.DomainHash, error)
	HasTips(dbContext DBReader) (bool, error)
}
