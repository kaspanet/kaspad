package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SyncManager exposes functions to support sync between kaspad nodes
type SyncManager interface {
	GetHashesBetween(lowHash, highHash *externalapi.DomainHash, maxBlueScoreDifference uint64) ([]*externalapi.DomainHash, error)
	GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error)
	CreateHeadersSelectedChainBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error)
	GetSyncInfo() (*externalapi.SyncInfo, error)
}
