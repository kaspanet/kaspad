package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte, transactions []*model.DomainTransaction) (*model.DomainBlock, error)
	ValidateAndInsertBlock(block *model.DomainBlock) error
	ValidateTransactionAndPopulateWithConsensusData(transaction *model.DomainTransaction) error
}

type consensus struct {
	blockProcessor        model.BlockProcessor
	consensusStateManager model.ConsensusStateManager
	transactionValidator  model.TransactionValidator
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (s *consensus) BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	transactions []*model.DomainTransaction) (*model.DomainBlock, error) {

	return s.blockProcessor.BuildBlock(coinbaseScriptPublicKey, coinbaseExtraData, transactions)
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (s *consensus) ValidateAndInsertBlock(block *model.DomainBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
}

// ValidateTransactionAndPopulateWithConsensusData validates the given transaction
// and populates it with any missing consensus data
func (s *consensus) ValidateTransactionAndPopulateWithConsensusData(transaction *model.DomainTransaction) error {
	return s.transactionValidator.ValidateTransactionAndPopulateWithConsensusData(transaction)
}
