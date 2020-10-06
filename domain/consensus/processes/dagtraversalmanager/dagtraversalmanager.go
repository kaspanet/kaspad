package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes"
	"github.com/kaspanet/kaspad/util/daghash"
)

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager struct {
	dagTopologyManager processes.DAGTopologyManager
	ghostdagManager    processes.GHOSTDAGManager
}

// New instantiates a new DAGTraversalManager
func New(
	dagTopologyManager processes.DAGTopologyManager,
	ghostdagManager processes.GHOSTDAGManager) *DAGTraversalManager {
	return &DAGTraversalManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagManager:    ghostdagManager,
	}
}

// BlockAtDepth returns the hash of the block that's at the
// given depth from the given highHash
func (dtm *DAGTraversalManager) BlockAtDepth(highHash *daghash.Hash, depth uint64) *daghash.Hash {
	return nil
}

// SelectedParentIterator creates an iterator over the selected
// parent chain of the given highHash
func (dtm *DAGTraversalManager) SelectedParentIterator(highHash *daghash.Hash) model.SelectedParentIterator {
	return nil
}
