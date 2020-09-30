package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningManager ...
type PruningManager struct {
	dagTraversalManager algorithms.DAGTraversalManager
	pruningPointStore   datastructures.PruningPointStore
}

// New ...
func New(
	dagTraversalManager algorithms.DAGTraversalManager,
	pruningPointStore datastructures.PruningPointStore) *PruningManager {
	return &PruningManager{
		dagTraversalManager: dagTraversalManager,
		pruningPointStore:   pruningPointStore,
	}
}

// FindPruningPoint ...
func (pm *PruningManager) FindPruningPoint(blockHash *daghash.Hash) *daghash.Hash {
	return nil
}
