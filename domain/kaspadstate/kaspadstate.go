package kaspadstate

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util"
)

// KaspadState maintains the current core state of the node
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

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (s *kaspadState) BuildBlock(transactionSelector model.TransactionSelector) *appmessage.MsgBlock {
	return s.blockProcessor.BuildBlock(transactionSelector)
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (s *kaspadState) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
}

// UTXOByOutpoint returns a UTXOEntry matching the given outpoint
func (s *kaspadState) UTXOByOutpoint(outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return s.consensusStateManager.UTXOByOutpoint(outpoint)
}

// ValidateTransaction validates the given transaction against the
// current state using the given utxoEntries
func (s *kaspadState) ValidateTransaction(transaction *util.Tx, utxoEntries []*model.UTXOEntry) error {
	return s.consensusStateManager.ValidateTransaction(transaction, utxoEntries)
}
