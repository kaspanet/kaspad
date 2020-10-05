package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/datastructures"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes"
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningManager resolves and manages the current pruning point
type PruningManager struct {
	dagTraversalManager processes.DAGTraversalManager
	pruningPointStore   datastructures.PruningPointStore
}

// New instantiates a new PruningManager
func New(
	dagTraversalManager processes.DAGTraversalManager,
	pruningPointStore datastructures.PruningPointStore) *PruningManager {
	return &PruningManager{
		dagTraversalManager: dagTraversalManager,
		pruningPointStore:   pruningPointStore,
	}
}

// FindNextPruningPoint ...
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
