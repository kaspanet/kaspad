package dagtraversalmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

type DAGTraversalManager struct {
	dagTopologyManager algorithms.DAGTopologyManager
	ghostdagManager    algorithms.GHOSTDAGManager
}

func New(
	dagTopologyManager algorithms.DAGTopologyManager,
	ghostdagManager algorithms.GHOSTDAGManager) *DAGTraversalManager {
	return &DAGTraversalManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagManager:    ghostdagManager,
	}
}

func (dtm *DAGTraversalManager) BlockAtDepth(uint64) *daghash.Hash {
	return nil
}

func (dtm *DAGTraversalManager) SelectedParentIterator(highHash *daghash.Hash) model.SelectedParentIterator {
	return nil
}
