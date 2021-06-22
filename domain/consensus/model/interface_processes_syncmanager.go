package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SyncManager exposes functions to support sync between kaspad nodes
type SyncManager interface {
	GetHashesBetween(stagingArea *StagingArea, lowHash, highHash *externalapi.DomainHash, maxBlueScoreDifference uint64) (
		hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error)
	CreateBlockLocator(stagingArea *StagingArea, lowHash, highHash *externalapi.DomainHash, limit uint32) (
		externalapi.BlockLocator, error)
	CreateHeadersSelectedChainBlockLocator(stagingArea *StagingArea, lowHash, highHash *externalapi.DomainHash) (
		externalapi.BlockLocator, error)
	GetSyncInfo(stagingArea *StagingArea) (*externalapi.SyncInfo, error)
}
