package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator interface {
	ValidateHeaderInIsolation(block *externalapi.DomainBlock) error
	ValidateBodyInIsolation(block *externalapi.DomainBlock) error
	ValidateHeaderInContext(block *externalapi.DomainBlock) error
	ValidateBodyInContext(block *externalapi.DomainBlock) error
	ValidateAgainstPastUTXO(block *externalapi.DomainBlock) error
	ValidateFinality(block *externalapi.DomainBlock) error
}
