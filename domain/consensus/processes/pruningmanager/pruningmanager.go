package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
	databaseContext model.DBReader

	dagTraversalManager   model.DAGTraversalManager
	dagTopologyManager    model.DAGTopologyManager
	consensusStateManager model.ConsensusStateManager
	ghostdagDataStore     model.GHOSTDAGDataStore
	pruningStore          model.PruningStore
	blockStatusStore      model.BlockStatusStore

	multiSetStore       model.MultisetStore
	acceptanceDataStore model.AcceptanceDataStore
	blocksStore         model.BlockStore
	utxoDiffStore       model.UTXODiffStore

	pruningDepth     uint64
	finalityInterval uint64
}

// New instantiates a new PruningManager
func New(
	databaseContext model.DBReader,

	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	consensusStateManager model.ConsensusStateManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	pruningStore model.PruningStore,
	blockStatusStore model.BlockStatusStore,

	multiSetStore model.MultisetStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blocksStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,

	finalityInterval uint64,
	k model.KType,
) model.PruningManager {

	return &pruningManager{
		databaseContext:       databaseContext,
		dagTraversalManager:   dagTraversalManager,
		dagTopologyManager:    dagTopologyManager,
		consensusStateManager: consensusStateManager,
		ghostdagDataStore:     ghostdagDataStore,
		pruningStore:          pruningStore,
		blockStatusStore:      blockStatusStore,
		multiSetStore:         multiSetStore,
		acceptanceDataStore:   acceptanceDataStore,
		blocksStore:           blocksStore,
		utxoDiffStore:         utxoDiffStore,
		pruningDepth:          pruningDepth(uint64(k), finalityInterval, constants.MergeSetSizeLimit),
		finalityInterval:      finalityInterval,
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
