package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SyncManager exposes functions to support sync between kaspad nodes
type SyncManager interface {
	GetHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	CreateBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error)
	FindNextBlockLocatorBoundaries(blockLocator externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error)
	IsBlockInHeaderPruningPointFutureAndVirtualPast(blockHash *externalapi.DomainHash) (bool, error)
}
