package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeadersSelectedTipManager manages the state of the headers selected tip
type HeadersSelectedTipManager interface {
	AddHeaderTip(hash *externalapi.DomainHash) error
}
