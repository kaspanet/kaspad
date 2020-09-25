package pruningmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/state/datastructures/pruningpointstore"
)

type PruningManager struct {
	dagTraversalManager dagtraversalmanager.DAGTraversalManager
	pruningPointStore   pruningpointstore.PruningPointStore
}

func New(
	dagTraversalManager dagtraversalmanager.DAGTraversalManager,
	pruningPointStore pruningpointstore.PruningPointStore) *PruningManager {
	return &PruningManager{
		dagTraversalManager: dagTraversalManager,
		pruningPointStore:   pruningPointStore,
	}
}
