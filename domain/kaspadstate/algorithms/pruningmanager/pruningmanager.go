package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
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
func (pm *PruningManager) FindNextPruningPoint(blockHash *daghash.Hash) (found bool,
	newPruningPoint *daghash.Hash, newPruningPointUTXOSet model.ReadOnlyUTXOSet) {

	return false, nil, nil
}

// PruningPoint ...
func (pm *PruningManager) PruningPoint() *daghash.Hash {
	return nil
}

// SerializedUTXOSet ...
func (pm *PruningManager) SerializedUTXOSet() []byte {
	return nil
}
