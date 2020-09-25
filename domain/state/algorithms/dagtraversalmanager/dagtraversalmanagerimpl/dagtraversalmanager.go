package dagtraversalmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/state/algorithms/ghostdagmanager"
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
