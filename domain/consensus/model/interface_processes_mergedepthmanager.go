package model

import "github.com/c4ei/YunSeokYeol/domain/consensus/model/externalapi"

// MergeDepthManager is used to validate mergeDepth for blocks
type MergeDepthManager interface {
	CheckBoundedMergeDepth(stagingArea *StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error
	NonBoundedMergeDepthViolatingBlues(stagingArea *StagingArea, blockHash, mergeDepthRoot *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	VirtualMergeDepthRoot(stagingArea *StagingArea) (*externalapi.DomainHash, error)
}
