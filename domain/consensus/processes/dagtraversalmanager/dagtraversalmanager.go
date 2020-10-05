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

// BlockAtDepth ...
func (dtm *DAGTraversalManager) BlockAtDepth(uint64) *daghash.Hash {
	return nil
}

// SelectedParentIterator ...
func (dtm *DAGTraversalManager) SelectedParentIterator(highHash *daghash.Hash) model.SelectedParentIterator {
	return nil
}
