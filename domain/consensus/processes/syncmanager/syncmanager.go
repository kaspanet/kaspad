package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type syncManager struct {
	dagTraversalManager model.DAGTraversalManager
}

// New instantiates a new SyncManager
func New(dagTraversalManager model.DAGTraversalManager) model.SyncManager {
	return &syncManager{
		dagTraversalManager: dagTraversalManager,
	}
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
