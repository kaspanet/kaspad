package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
)

type PruningManager struct {
	dagTraversalManager algorithms.DAGTraversalManager
	pruningPointStore   datastructures.PruningPointStore
}

func New(
	dagTraversalManager algorithms.DAGTraversalManager,
	pruningPointStore datastructures.PruningPointStore) *PruningManager {
	return &PruningManager{
		dagTraversalManager: dagTraversalManager,
		pruningPointStore:   pruningPointStore,
	}
}

func (pm *PruningManager) UpdatePruningPointAndPruneIfRequired() {

}
