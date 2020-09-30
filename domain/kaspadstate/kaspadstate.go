package kaspadstate

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util"
)

// KaspadState ...
type KaspadState interface {
	BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error

	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error
}

type kaspadState struct {
	blockProcessor        algorithms.BlockProcessor
	consensusStateManager algorithms.ConsensusStateManager
}

func (s *kaspadState) BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock {
	return s.blockProcessor.BuildBlock(transactionSelector)
}

func (s *kaspadState) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
}

func (s *kaspadState) UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return s.consensusStateManager.UTXOByOutpoint(outpoint)
}

func (s *kaspadState) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return s.consensusStateManager.ValidateTransaction(transaction, utxoEntries)
}
