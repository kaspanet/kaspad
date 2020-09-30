package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util"
)

type ConsensusStateManager struct {
	dagParams *dagconfig.Params

	consensusStateStore datastructures.ConsensusStateStore
	multisetStore       datastructures.MultisetStore
	utxoDiffStore       datastructures.UTXODiffStore
}

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

func (csm *ConsensusStateManager) UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return nil
}

func (csm *ConsensusStateManager) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return nil
}

func (csm *ConsensusStateManager) SerializedUTXOSet() []byte {
	return nil
}

func (csm *ConsensusStateManager) UpdateConsensusState(block *appmessage.MsgBlock) {
}

func (csm *ConsensusStateManager) ValidateBlockTransactions(block *appmessage.MsgBlock) error {
	return nil
}
