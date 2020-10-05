package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util"
)

// ConsensusStateManager ...
type ConsensusStateManager struct {
	dagParams *dagconfig.Params

	consensusStateStore datastructures.ConsensusStateStore
	multisetStore       datastructures.MultisetStore
	utxoDiffStore       datastructures.UTXODiffStore
}

// New ...
func New(
	dagParams *dagconfig.Params,
	consensusStateStore datastructures.ConsensusStateStore,
	multisetStore datastructures.MultisetStore,
	utxoDiffStore datastructures.UTXODiffStore) *ConsensusStateManager {
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

// ValidateTransaction validates the given transaction against the
// current state using the given utxoEntries
func (csm *ConsensusStateManager) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return nil
}

// CalculateConsensusStateChanges ...
func (csm *ConsensusStateManager) CalculateConsensusStateChanges(block *appmessage.MsgBlock) *model.ConsensusStateChanges {
	return nil
}
