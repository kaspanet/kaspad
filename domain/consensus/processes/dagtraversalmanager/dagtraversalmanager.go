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

// New instantiates a new dagTraversalManager
func New(
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager) model.DAGTraversalManager {
	return &dagTraversalManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagManager:    ghostdagManager,
	}
}

// BlockAtDepth returns the hash of the block that's at the
// given depth from the given highHash
func (dtm *dagTraversalManager) BlockAtDepth(highHash *model.DomainHash, depth uint64) *model.DomainHash {
	return nil
}

// SelectedParentIterator creates an iterator over the selected
// parent chain of the given highHash
func (dtm *dagTraversalManager) SelectedParentIterator(highHash *model.DomainHash) model.SelectedParentIterator {
	return nil
}

// ChainBlockAtBlueScore returns the hash of the smallest block
// with a blue score greater than the given blueScore in the
// virtual block's selected parent chain
func (dtm *dagTraversalManager) ChainBlockAtBlueScore(blueScore uint64) *model.DomainHash {
	return nil
}
