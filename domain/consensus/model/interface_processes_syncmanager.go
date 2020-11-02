package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SyncManager exposes functions to support sync between kaspad nodes
type SyncManager interface {
	GetHashesBetween(lowHigh, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	CreateBlockLocator(lowHigh, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error)
	FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHigh, highHash *externalapi.DomainHash, err error)
	IsBlockHeaderInPruningPointFutureAndVirtualPast(blockHash *externalapi.DomainHash) (bool, error)
}
