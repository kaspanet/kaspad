package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Consensus maintains the current core state of the node
type Consensus interface {
	BuildBlock(coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error)
	ValidateAndInsertBlock(block *externalapi.DomainBlock) error
	ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error

	GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error)
	GetBlockHeader(blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error)
	GetBlockInfo(blockHash *externalapi.DomainHash) (*externalapi.BlockInfo, error)

	GetHashesBetween(lowHigh, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	GetHashesAbovePruningPoint(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	GetPruningPointUTXOSet() ([]byte, error)
	SetPruningPointUTXOSet(pruningPoint *externalapi.DomainHash, serializedUTXOSet []byte) error
	GetVirtualSelectedParent() (*externalapi.DomainBlock, error)
	CreateBlockLocator(lowHigh, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error)
	FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHigh, highHash *externalapi.DomainHash, err error)
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
func (s *consensus) ValidateAndInsertBlock(block *externalapi.DomainBlock) error {
	return s.blockProcessor.ValidateAndInsertBlock(block)
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

func (s *consensus) GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	panic("implement me")
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	panic("implement me")
}

func (s *consensus) GetBlockInfo(blockHash *externalapi.DomainHash) (*externalapi.BlockInfo, error) {
	panic("implement me")
}

func (s *consensus) GetHashesBetween(lowHigh, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (s *consensus) GetHashesAbovePruningPoint(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (s *consensus) GetPruningPointUTXOSet() ([]byte, error) {
	panic("implement me")
}

func (s *consensus) SetPruningPointUTXOSet(pruningPoint *externalapi.DomainHash, serializedUTXOSet []byte) error {
	panic("implement me")
}

func (s *consensus) GetVirtualSelectedParent() (*externalapi.DomainBlock, error) {
	panic("implement me")
}

func (s *consensus) CreateBlockLocator(lowHigh, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error) {
	panic("implement me")
}

func (s *consensus) FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHigh, highHash *externalapi.DomainHash, err error) {
	panic("implement me")
}
