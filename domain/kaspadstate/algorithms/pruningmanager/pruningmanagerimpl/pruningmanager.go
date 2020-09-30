package pruningmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
)

type PruningManager struct {
	dagTraversalManager dagtraversalmanager.DAGTraversalManager
	pruningPointStore   datastructures.PruningPointStore
}

func New(
	dagTraversalManager dagtraversalmanager.DAGTraversalManager,
	pruningPointStore datastructures.PruningPointStore) *PruningManager {
	return &PruningManager{
		dagTraversalManager: dagTraversalManager,
		pruningPointStore:   pruningPointStore,
	}
}

func (pm *PruningManager) UpdatePruningPointAndPruneIfRequired() {

}
