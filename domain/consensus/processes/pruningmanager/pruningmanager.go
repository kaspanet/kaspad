package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
	dagTraversalManager   model.DAGTraversalManager
	pruningStore          model.PruningStore
	dagTopologyManager    model.DAGTopologyManager
	blockStatusStore      model.BlockStatusStore
	consensusStateManager model.ConsensusStateManager
}

// New instantiates a new pruningManager
func New(
	dagTraversalManager model.DAGTraversalManager,
	pruningStore model.PruningStore,
	dagTopologyManager model.DAGTopologyManager,
	blockStatusStore model.BlockStatusStore,
	consensusStateManager model.ConsensusStateManager) model.PruningManager {
	return &pruningManager{
		dagTraversalManager:   dagTraversalManager,
		pruningStore:          pruningStore,
		dagTopologyManager:    dagTopologyManager,
		blockStatusStore:      blockStatusStore,
		consensusStateManager: consensusStateManager,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash. If none found, returns false
func (pm *pruningManager) FindNextPruningPoint(blockGHOSTDAGData *model.BlockGHOSTDAGData) (found bool,
	newPruningPoint *model.DomainHash, newPruningPointUTXOSet model.ReadOnlyUTXOSet) {

	return false, nil, nil
}

// PruningPoint returns the hash of the current pruning point
func (pm *pruningManager) PruningPoint() *model.DomainHash {
	return nil
}

// SerializedUTXOSet returns the serialized UTXO set of the
// current pruning point
func (pm *pruningManager) SerializedUTXOSet() []byte {
	return nil
}
