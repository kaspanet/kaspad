package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// HeaderTipsManager manages the state of the header tips
type HeaderTipsManager interface {
	AddBlock(hash *externalapi.DomainHash) error
}
