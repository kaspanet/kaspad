package consensus

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

type consensus struct {
	lock            *sync.Mutex
	databaseContext model.DBReader

	blockProcessor        model.BlockProcessor
	blockBuilder          model.BlockBuilder
	consensusStateManager model.ConsensusStateManager
	transactionValidator  model.TransactionValidator
	syncManager           model.SyncManager
	pastMedianTimeManager model.PastMedianTimeManager
	blockValidator        model.BlockValidator
	coinbaseManager       model.CoinbaseManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	difficultyManager     model.DifficultyManager
	ghostdagManager       model.GHOSTDAGManager
	headerTipsManager     model.HeaderTipsManager
	mergeDepthManager     model.MergeDepthManager
	pruningManager        model.PruningManager
	reachabilityManager   model.ReachabilityManager
	finalityManager       model.FinalityManager

	acceptanceDataStore   model.AcceptanceDataStore
	blockStore            model.BlockStore
	blockHeaderStore      model.BlockHeaderStore
	pruningStore          model.PruningStore
	ghostdagDataStore     model.GHOSTDAGDataStore
	blockRelationStore    model.BlockRelationStore
	blockStatusStore      model.BlockStatusStore
	consensusStateStore   model.ConsensusStateStore
	headerTipsStore       model.HeaderTipsStore
	multisetStore         model.MultisetStore
	reachabilityDataStore model.ReachabilityDataStore
	utxoDiffStore         model.UTXODiffStore
	finalityStore         model.FinalityStore
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (s *consensus) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockBuilder.BuildBlock(coinbaseData, transactions)
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (s *consensus) ValidateAndInsertBlock(block *externalapi.DomainBlock) (*externalapi.InsertBlockResult, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertBlock(block)
}

// ValidateTransactionAndPopulateWithConsensusData validates the given transaction
// and populates it with any missing consensus data
func (s *consensus) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.transactionValidator.ValidateTransactionInIsolation(transaction)
	if err != nil {
		return err
	}

	err = s.consensusStateManager.PopulateTransactionWithUTXOEntries(transaction)
	if err != nil {
		return err
	}

	virtualSelectedParentMedianTime, err := s.pastMedianTimeManager.PastMedianTime(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	return s.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(transaction,
		model.VirtualBlockHash, virtualSelectedParentMedianTime)
}

func (s *consensus) GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockStore.Block(s.databaseContext, blockHash)
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockHeaderStore.BlockHeader(s.databaseContext, blockHash)
}

func (s *consensus) GetBlockInfo(blockHash *externalapi.DomainHash) (*externalapi.BlockInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	blockInfo := &externalapi.BlockInfo{}

	exists, err := s.blockStatusStore.Exists(s.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.Exists = exists
	if !exists {
		return blockInfo, nil
	}

	blockStatus, err := s.blockStatusStore.Get(s.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.BlockStatus = blockStatus

	// If the status is invalid, then we don't have the necessary reachability data to check if it's in PruningPoint.Future.
	if blockStatus == externalapi.StatusInvalid {
		return blockInfo, nil
	}

	ghostdagData, err := s.ghostdagDataStore.Get(s.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	blockInfo.BlueScore = ghostdagData.BlueScore()

	isBlockInHeaderPruningPointFuture, err := s.syncManager.IsBlockInHeaderPruningPointFuture(blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.IsBlockInHeaderPruningPointFuture = isBlockInHeaderPruningPointFuture

	return blockInfo, nil
}

func (s *consensus) GetHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.GetHashesBetween(lowHash, highHash)
}

func (s *consensus) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.GetMissingBlockBodyHashes(highHash)
}

func (s *consensus) GetPruningPointUTXOSet(expectedPruningPointHash *externalapi.DomainHash) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	pruningPointHash, err := s.pruningStore.PruningPoint(s.databaseContext)
	if err != nil {
		return nil, err
	}

	if *expectedPruningPointHash != *pruningPointHash {
		return nil, errors.Wrapf(ruleerrors.ErrWrongPruningPointHash, "expected pruning point %s but got %s",
			expectedPruningPointHash,
			pruningPointHash)
	}

	serializedUTXOSet, err := s.pruningStore.PruningPointSerializedUTXOSet(s.databaseContext)
	if err != nil {
		return nil, err
	}
	return serializedUTXOSet, nil
}

func (s *consensus) SetPruningPointUTXOSet(serializedUTXOSet []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.consensusStateManager.SetPruningPointUTXOSet(serializedUTXOSet)
}

func (s *consensus) GetVirtualSelectedParent() (*externalapi.DomainBlock, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	virtualGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	return s.blockStore.Block(s.databaseContext, virtualGHOSTDAGData.SelectedParent())
}

func (s *consensus) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.CreateBlockLocator(lowHash, highHash, limit)
}

func (s *consensus) FindNextBlockLocatorBoundaries(blockLocator externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.FindNextBlockLocatorBoundaries(blockLocator)
}

func (s *consensus) GetSyncInfo() (*externalapi.SyncInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.GetSyncInfo()
}
