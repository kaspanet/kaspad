package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	dagParams *dagconfig.Params

	databaseContext     *database.DomainDBContext
	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	utxoDiffStore       model.UTXODiffStore
	blockStore          model.BlockStore
}

// New instantiates a new consensusStateManager
func New(
	databaseContext *database.DomainDBContext,
	dagParams *dagconfig.Params,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	utxoDiffStore model.UTXODiffStore,
	blockStore model.BlockStore) model.ConsensusStateManager {

	return &consensusStateManager{
		dagParams: dagParams,

		databaseContext:     databaseContext,
		consensusStateStore: consensusStateStore,
		multisetStore:       multisetStore,
		utxoDiffStore:       utxoDiffStore,
		blockStore:          blockStore,
	}
}

// UTXOByOutpoint returns a UTXOEntry matching the given outpoint
func (csm *consensusStateManager) UTXOByOutpoint(outpoint *model.DomainOutpoint) *model.UTXOEntry {
	return nil
}

// CalculateConsensusStateChanges returns a set of changes that must occur in order
// to transition the current consensus state into the one including the given block
func (csm *consensusStateManager) CalculateConsensusStateChanges(block *model.DomainBlock, isDisqualified bool) (
	stateChanges *model.ConsensusStateChanges, utxoDiffChanges *model.UTXODiffChanges,
	virtualGHOSTDAGData *model.BlockGHOSTDAGData) {

	return nil, nil, nil
}

// CalculateAcceptanceDataAndUTXOMultiset calculates and returns the acceptance data and the
// multiset associated with the given blockHash
func (csm *consensusStateManager) CalculateAcceptanceDataAndUTXOMultiset(blockGHOSTDAGData *model.BlockGHOSTDAGData) (
	*model.BlockAcceptanceData, model.Multiset) {

	return nil, nil
}

// Tips returns the current DAG tips
func (csm *consensusStateManager) Tips() []*model.DomainHash {
	return nil
}

// VirtualData returns the medianTime and blueScore of the current virtual block
func (csm *consensusStateManager) VirtualData() (medianTime int64, blueScore uint64) {
	return 0, 0
}

// RestoreUTXOSet calculates and returns the UTXOSet of the given blockHash
func (csm *consensusStateManager) RestorePastUTXOSet(blockHash *model.DomainHash) model.ReadOnlyUTXOSet {
	return nil
}
