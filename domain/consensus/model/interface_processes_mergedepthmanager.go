package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type MergeDepthManager interface {
	CheckBoundedMergeDepth(blockHash *externalapi.DomainHash) error
	NonBoundedMergeDepthViolatingBlues(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
}
