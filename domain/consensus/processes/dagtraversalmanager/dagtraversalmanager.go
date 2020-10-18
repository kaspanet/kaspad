package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// dagTraversalManager exposes methods for travering blocks
// in the DAG
type dagTraversalManager struct {
	dagTopologyManager model.DAGTopologyManager
	ghostdagManager    model.GHOSTDAGManager
}

// New instantiates a new DAGTraversalManager
func New(
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager) model.DAGTraversalManager {
	return &dagTraversalManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagManager:    ghostdagManager,
	}
}

// SelectedParentIterator creates an iterator over the selected
// parent chain of the given highHash
func (dtm *dagTraversalManager) SelectedParentIterator(highHash *model.DomainHash) (model.SelectedParentIterator, error) {
	return nil, nil
}

// HighestChainBlockBelowBlueScore returns the hash of the
// highest block with a blue score lower than the given
// blueScore in the block with the given highHash's selected
// parent chain
func (dtm *dagTraversalManager) HighestChainBlockBelowBlueScore(highHash *model.DomainHash, blueScore uint64) (*model.DomainHash, error) {
	return nil, nil
}
