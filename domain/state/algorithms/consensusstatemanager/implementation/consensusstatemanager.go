package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/util"
)

type ConsensusStateManager struct {
}

func New() *ConsensusStateManager {
	return &ConsensusStateManager{}
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
