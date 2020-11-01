package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)
	ValidateAndInsertBlock(block *externalapi.DomainBlock, headerOnly bool) error
	ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error
}

type consensus struct {
	blockProcessor        model.BlockProcessor
	consensusStateManager model.ConsensusStateManager
	transactionValidator  model.TransactionValidator
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (s *consensus) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	return s.blockProcessor.BuildBlock(coinbaseData, transactions)
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (s *consensus) ValidateAndInsertBlock(block *externalapi.DomainBlock, headerOnly bool) error {
	return s.blockProcessor.ValidateAndInsertBlock(block, headerOnly)
}

// ValidateTransactionAndPopulateWithConsensusData validates the given transaction
// and populates it with any missing consensus data
func (s *consensus) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	err := s.transactionValidator.ValidateTransactionInIsolation(transaction)
	if err != nil {
		return err
	}

	err = s.consensusStateManager.PopulateTransactionWithUTXOEntries(transaction)
	if err != nil {
		return err
	}

	return s.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(transaction,
		validateTransactionInContextAndPopulateMassAndFeeVirtualBlockHash(),
		validateTransactionInContextAndPopulateMassAndFeeSelectedParentMedianTime())
}

func validateTransactionInContextAndPopulateMassAndFeeSelectedParentMedianTime() int64 {
	panic("unimplemented")
}

func validateTransactionInContextAndPopulateMassAndFeeVirtualBlockHash() *externalapi.DomainHash {
	panic("unimplemented")
}
