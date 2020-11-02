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

func (s syncManager) GetHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (s syncManager) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (s syncManager) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error) {
	panic("implement me")
}

func (s syncManager) FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	panic("implement me")
}

func (s syncManager) IsBlockHeaderInPruningPointFutureAndVirtualPast(blockHash *externalapi.DomainHash) (bool, error) {
	panic("implement me")
}
