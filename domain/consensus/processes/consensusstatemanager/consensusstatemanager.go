package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager struct {
	dagParams *dagconfig.Params

	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	utxoDiffStore       model.UTXODiffStore
	blockStore          model.BlockStore
}

// New instantiates a new ConsensusStateManager
func New(
	dagParams *dagconfig.Params,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	utxoDiffStore model.UTXODiffStore,
	blockStore model.BlockStore) *ConsensusStateManager {
	return &ConsensusStateManager{
		dagParams: dagParams,

		consensusStateStore: consensusStateStore,
		multisetStore:       multisetStore,
		utxoDiffStore:       utxoDiffStore,
		blockStore:          blockStore,
	}
}

// UTXOByOutpoint returns a UTXOEntry matching the given outpoint
func (csm *ConsensusStateManager) UTXOByOutpoint(outpoint *model.DomainOutpoint) *model.UTXOEntry {
	return nil
}

// CalculateConsensusStateChanges returns a set of changes that must occur in order
// to transition the current consensus state into the one including the given block
func (csm *ConsensusStateManager) CalculateConsensusStateChanges(block *model.DomainBlock, parents []*model.DomainHash, transactions []*model.DomainTransaction, isDisqualified bool) (stateChanges *model.ConsensusStateChanges, utxoDiffChanges *model.UTXODiffChanges, virtualGHOSTDAGData *model.BlockGHOSTDAGData) {
	return nil, nil, nil
}

// CalculateAcceptanceDataAndMultiset calculates and returns the acceptance data and the
// multiset associated with the given blockHash
func (csm *ConsensusStateManager) CalculateAcceptanceDataAndMultiset(blockHash *model.DomainHash) (*model.BlockAcceptanceData, model.Multiset) {
	return nil, nil
}

// Tips returns the current DAG tips
func (csm *ConsensusStateManager) Tips() []*model.DomainHash {
	return nil
}

// VirtualData returns the medianTime and blueScore of the current virtual block
func (csm *ConsensusStateManager) VirtualData() (medianTime int64, blueScore uint64) {
	return 0, 0
}

// RestoreUTXOSet calculates and returns the UTXOSet of the given blockHash
func (csm *ConsensusStateManager) RestoreUTXOSet(blockHash *model.DomainHash) model.ReadOnlyUTXOSet {
	return nil
}
