package dagtraversalmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/ghostdagmanager"
	"github.com/kaspanet/kaspad/util/daghash"
)

type DAGTraversalManager struct {
	dagTopologyManager dagtopologymanager.DAGTopologyManager
	ghostdagManager    ghostdagmanager.GHOSTDAGManager
}

func New(
	dagTopologyManager dagtopologymanager.DAGTopologyManager,
	ghostdagManager ghostdagmanager.GHOSTDAGManager) *DAGTraversalManager {
	return &DAGTraversalManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagManager:    ghostdagManager,
	}
}

func (dtm *DAGTraversalManager) BlockAtDepth(uint64) *daghash.Hash {
	return nil
}

func (dtm *DAGTraversalManager) SelectedParentIterator(highHash *daghash.Hash) dagtraversalmanager.SelectedParentIterator {
	return nil
}
