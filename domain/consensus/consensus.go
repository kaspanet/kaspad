package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(scriptPublicKey []byte, extraData []byte, transactionSelector model.TransactionSelector) *model.DomainBlock
	ValidateAndInsertBlock(block *model.DomainBlock) error
	UTXOByOutpoint(outpoint *model.DomainOutpoint) *model.UTXOEntry
	ValidateTransaction(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) error
}

type consensus struct {
	blockProcessor        model.BlockProcessor
	consensusStateManager model.ConsensusStateManager
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (s *consensus) BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	transactionSelector model.TransactionSelector) *model.DomainBlock {

	return s.blockProcessor.BuildBlock(coinbaseScriptPublicKey, coinbaseExtraData, transactionSelector)
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (s *consensus) ValidateAndInsertBlock(block *model.DomainBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
}

// UTXOByOutpoint returns a UTXOEntry matching the given outpoint
func (s *consensus) UTXOByOutpoint(outpoint *model.DomainOutpoint) *model.UTXOEntry {
	return s.consensusStateManager.UTXOByOutpoint(outpoint)
}

// ValidateTransaction validates the given transaction using
// the given utxoEntries
func (s *consensus) ValidateTransaction(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) error {
	return s.consensusStateManager.ValidateTransaction(transaction, utxoEntries)
}
