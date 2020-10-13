package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager struct {
	dagTopologyManager model.DAGTopologyManager
	ghostdagManager    model.GHOSTDAGManager
}

// New instantiates a new DAGTraversalManager
func New(
	dagTopologyManager model.DAGTopologyManager,
	ghostdagManager model.GHOSTDAGManager) *DAGTraversalManager {
	return &DAGTraversalManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagManager:    ghostdagManager,
	}
}

// BlockAtDepth returns the hash of the block that's at the
// given depth from the given highHash
func (dtm *DAGTraversalManager) BlockAtDepth(highHash *model.DomainHash, depth uint64) *model.DomainHash {
	return nil
}

// SelectedParentIterator creates an iterator over the selected
// parent chain of the given highHash
func (dtm *DAGTraversalManager) SelectedParentIterator(highHash *model.DomainHash) model.SelectedParentIterator {
	return nil
}

// ChainBlockAtBlueScore returns the hash of the smallest block
// with a blue score greater than the given blueScore in the
// virtual block's selected parent chain
func (dtm *DAGTraversalManager) ChainBlockAtBlueScore(blueScore uint64) *model.DomainHash {
	return nil
}
