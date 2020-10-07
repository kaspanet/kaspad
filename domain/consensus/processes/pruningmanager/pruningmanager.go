package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// PruningManager resolves and manages the current pruning point
type PruningManager struct {
	dagTraversalManager model.DAGTraversalManager
	pruningPointStore   model.PruningPointStore
}

// New instantiates a new PruningManager
func New(
	dagTraversalManager model.DAGTraversalManager,
	pruningPointStore model.PruningPointStore) *PruningManager {
	return &PruningManager{
		dagTraversalManager: dagTraversalManager,
		pruningPointStore:   pruningPointStore,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash. If none found, returns false
func (pm *PruningManager) FindNextPruningPoint(blockHash *daghash.Hash) (found bool,
	newPruningPoint *daghash.Hash, newPruningPointUTXOSet model.ReadOnlyUTXOSet) {

	return false, nil, nil
}

// PruningPoint returns the hash of the current pruning point
func (pm *PruningManager) PruningPoint() *daghash.Hash {
	return nil
}

// SerializedUTXOSet returns the serialized UTXO set of the
// current pruning point
func (pm *PruningManager) SerializedUTXOSet() []byte {
	return nil
}
