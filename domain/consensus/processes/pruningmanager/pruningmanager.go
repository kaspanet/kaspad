package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// PruningManager resolves and manages the current pruning point
type PruningManager struct {
	dagTraversalManager   model.DAGTraversalManager
	pruningStore          model.PruningStore
	dagTopologyManager    model.DAGTopologyManager
	blockStatusStore      model.BlockStatusStore
	consensusStateManager model.ConsensusStateManager
}

// New instantiates a new PruningManager
func New(
	dagTraversalManager model.DAGTraversalManager,
	pruningStore model.PruningStore,
	dagTopologyManager model.DAGTopologyManager,
	blockStatusStore model.BlockStatusStore,
	consensusStateManager model.ConsensusStateManager) *PruningManager {
	return &PruningManager{
		dagTraversalManager:   dagTraversalManager,
		pruningStore:          pruningStore,
		dagTopologyManager:    dagTopologyManager,
		blockStatusStore:      blockStatusStore,
		consensusStateManager: consensusStateManager,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash. If none found, returns false
func (pm *PruningManager) FindNextPruningPoint(blockGHOSTDAGData *model.BlockGHOSTDAGData) (found bool,
	newPruningPoint *model.DomainHash, newPruningPointUTXOSet model.ReadOnlyUTXOSet) {

	return false, nil, nil
}

// PruningPoint returns the hash of the current pruning point
func (pm *PruningManager) PruningPoint() *model.DomainHash {
	return nil
}

// SerializedUTXOSet returns the serialized UTXO set of the
// current pruning point
func (pm *PruningManager) SerializedUTXOSet() []byte {
	return nil
}
