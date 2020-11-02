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
	GetSelectedParent() (*externalapi.DomainBlock, error)
	CreateBlockLocator(lowHigh, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error)
	FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHigh, highHash *externalapi.DomainHash, err error)
}

type consensus struct {
	databaseContext model.DBReader

	blockProcessor        model.BlockProcessor
	consensusStateManager model.ConsensusStateManager
	transactionValidator  model.TransactionValidator

	blockStore        model.BlockStore
	blockHeaderStore  model.BlockHeaderStore
	pruningStore      model.PruningStore
	ghostdagDataStore model.GHOSTDAGDataStore
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
	return s.blockStore.Block(s.databaseContext, blockHash)
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	return s.blockHeaderStore.BlockHeader(s.databaseContext, blockHash)
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
	return s.pruningStore.PruningPointSerializedUTXOSet(s.databaseContext)
}

func (s *consensus) GetSelectedParent() (*externalapi.DomainBlock, error) {
	virtualGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	return s.GetBlock(virtualGHOSTDAGData.SelectedParent)
}

func (s *consensus) CreateBlockLocator(lowHigh, highHash *externalapi.DomainHash) (*externalapi.BlockLocator, error) {
	panic("implement me")
}

func (s *consensus) FindNextBlockLocatorBoundaries(blockLocator *externalapi.BlockLocator) (lowHigh, highHash *externalapi.DomainHash, err error) {
	panic("implement me")
}
