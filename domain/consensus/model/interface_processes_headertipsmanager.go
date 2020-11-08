package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeaderTipsManager manages the state of the header tips
type HeaderTipsManager interface {
	AddHeaderTip(hash *externalapi.DomainHash) error
	SelectedTip() (*externalapi.DomainHash, bool, error)
}
