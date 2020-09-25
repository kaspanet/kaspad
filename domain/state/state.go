package state

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/domain/state/algorithms/blockprocessor"
	"github.com/kaspanet/kaspad/domain/state/algorithms/consensusstatemanager"
	"github.com/kaspanet/kaspad/util"
)

type State interface {
	BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error

	UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry
	ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error
}

type state struct {
	blockProcessor        blockprocessor.BlockProcessor
	consensusStateManager consensusstatemanager.ConsensusStateManager
}

func (s *state) BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock {
	return s.blockProcessor.BuildBlock(transactionSelector)
}

func (s *state) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
}

func (s *state) UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return s.consensusStateManager.UTXOByOutpoint(outpoint)
}

func (s *state) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return s.consensusStateManager.ValidateTransaction(transaction, utxoEntries)
}
