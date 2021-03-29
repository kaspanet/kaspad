package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// MergeDepthManager is used to validate mergeDepth for blocks
type MergeDepthManager interface {
	CheckBoundedMergeDepth(stagingArea *StagingArea, blockHash *externalapi.DomainHash) error
	NonBoundedMergeDepthViolatingBlues(stagingArea *StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
}
