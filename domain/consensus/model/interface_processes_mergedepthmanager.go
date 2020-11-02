package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// MergeDepthManager is used to validate mergeDepth for blocks
type MergeDepthManager interface {
	CheckBoundedMergeDepth(blockHash *externalapi.DomainHash) error
	NonBoundedMergeDepthViolatingBlues(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
}
