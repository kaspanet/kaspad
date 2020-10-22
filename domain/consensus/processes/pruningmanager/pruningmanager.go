package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
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
// given blockHash
func (pm *pruningManager) FindNextPruningPoint(blockHash *externalapi.DomainHash) error {
	return nil
}

// PruningPoint returns the hash of the current pruning point
func (pm *pruningManager) PruningPoint() (*externalapi.DomainHash, error) {
	return nil, nil
}

// SerializedUTXOSet returns the serialized UTXO set of the
// current pruning point
func (pm *pruningManager) SerializedUTXOSet() ([]byte, error) {
	return nil, nil
}
