package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// DAGTraversalManager ...
type DAGTraversalManager struct {
	dagTopologyManager algorithms.DAGTopologyManager
	ghostdagManager    algorithms.GHOSTDAGManager
}

// New ...
func New(
	dagTopologyManager algorithms.DAGTopologyManager,
	ghostdagManager algorithms.GHOSTDAGManager) *DAGTraversalManager {
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
