package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type syncManager struct {
}

// New instantiates a new SyncManager
func New() model.SyncManager {
	return &syncManager{}
}

func (s syncManager) GetHashesBetween(lowHigh, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (s syncManager) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (s syncManager) CreateBlockLocator(lowHigh, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error) {
	panic("implement me")
}

func (s syncManager) FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHigh, highHash *externalapi.DomainHash, err error) {
	panic("implement me")
}
