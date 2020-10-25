package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
	dagTraversalManager model.DAGTraversalManager
	dagTopologyManager  model.DAGTopologyManager

	pruningStore        model.PruningStore
	blockStatusStore    model.BlockStatusStore
	consensusStateStore model.ConsensusStateStore
}

// New instantiates a new PruningManager
func New(
	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	pruningStore model.PruningStore,
	blockStatusStore model.BlockStatusStore,
	consensusStateStore model.ConsensusStateStore) model.PruningManager {

	return &pruningManager{
		dagTraversalManager: dagTraversalManager,
		dagTopologyManager:  dagTopologyManager,

		pruningStore:        pruningStore,
		blockStatusStore:    blockStatusStore,
		consensusStateStore: consensusStateStore,
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
