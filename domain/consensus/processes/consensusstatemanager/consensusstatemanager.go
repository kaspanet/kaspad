package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager struct {
	dagParams *dagconfig.Params

	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	utxoDiffStore       model.UTXODiffStore
}

// New instantiates a new ConsensusStateManager
func New(
	dagParams *dagconfig.Params,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	utxoDiffStore model.UTXODiffStore) *ConsensusStateManager {
	return &ConsensusStateManager{
		dagParams: dagParams,

		consensusStateStore: consensusStateStore,
		multisetStore:       multisetStore,
		utxoDiffStore:       utxoDiffStore,
	}
}

// UTXOByOutpoint returns a UTXOEntry matching the given outpoint
func (csm *ConsensusStateManager) UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return nil
}

// ValidateTransaction validates the given transaction using
// the given utxoEntries
func (csm *ConsensusStateManager) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return nil
}

// CalculateConsensusStateChanges returns a set of changes that must occur in order
// to transition the current consensus state into the one including the given block
func (csm *ConsensusStateManager) CalculateConsensusStateChanges(block *appmessage.MsgBlock) *model.ConsensusStateChanges {
	return nil
}
