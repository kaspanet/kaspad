package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util"
)

type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error

	SerializedUTXOSet() []byte
	UpdateConsensusState(block *appmessage.MsgBlock)
	ValidateBlockTransactions(block *appmessage.MsgBlock) error
}
