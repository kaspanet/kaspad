package consensus

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

type consensus struct {
	lock            *sync.Mutex
	databaseContext model.DBManager

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
	headerTipsManager     model.HeadersSelectedTipManager
	mergeDepthManager     model.MergeDepthManager
	pruningManager        model.PruningManager
	reachabilityManager   model.ReachabilityManager
	finalityManager       model.FinalityManager

	acceptanceDataStore       model.AcceptanceDataStore
	blockStore                model.BlockStore
	blockHeaderStore          model.BlockHeaderStore
	pruningStore              model.PruningStore
	ghostdagDataStore         model.GHOSTDAGDataStore
	blockRelationStore        model.BlockRelationStore
	blockStatusStore          model.BlockStatusStore
	consensusStateStore       model.ConsensusStateStore
	headersSelectedTipStore   model.HeaderSelectedTipStore
	multisetStore             model.MultisetStore
	reachabilityDataStore     model.ReachabilityDataStore
	utxoDiffStore             model.UTXODiffStore
	finalityStore             model.FinalityStore
	headersSelectedChainStore model.HeadersSelectedChainStore
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
func (s *consensus) ValidateAndInsertBlock(block *externalapi.DomainBlock) (*externalapi.BlockInsertionResult, error) {
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

	block, err := s.blockStore.Block(s.databaseContext, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, errors.Wrapf(err, "block %s does not exist", blockHash)
		}
		return nil, err
	}
	return block, nil
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	blockHeader, err := s.blockHeaderStore.BlockHeader(s.databaseContext, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, errors.Wrapf(err, "block header %s does not exist", blockHash)
		}
		return nil, err
	}
	return blockHeader, nil
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

	return blockInfo, nil
}

func (s *consensus) GetBlockAcceptanceData(blockHash *externalapi.DomainHash) (externalapi.AcceptanceData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(blockHash)
	if err != nil {
		return nil, err
	}

	return s.acceptanceDataStore.Get(s.databaseContext, blockHash)
}

func (s *consensus) GetHashesBetween(lowHash, highHash *externalapi.DomainHash,
	maxBlueScoreDifference uint64) ([]*externalapi.DomainHash, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(lowHash)
	if err != nil {
		return nil, err
	}
	err = s.validateBlockHashExists(highHash)
	if err != nil {
		return nil, err
	}

	return s.syncManager.GetHashesBetween(lowHash, highHash, maxBlueScoreDifference)
}

func (s *consensus) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(highHash)
	if err != nil {
		return nil, err
	}

	return s.syncManager.GetMissingBlockBodyHashes(highHash)
}

func (s *consensus) GetPruningPointUTXOs(expectedPruningPointHash *externalapi.DomainHash,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	pruningPointHash, err := s.pruningStore.PruningPoint(s.databaseContext)
	if err != nil {
		return nil, err
	}

	if !expectedPruningPointHash.Equal(pruningPointHash) {
		return nil, errors.Wrapf(ruleerrors.ErrWrongPruningPointHash, "expected pruning point %s but got %s",
			expectedPruningPointHash,
			pruningPointHash)
	}

	pruningPointUTXOs, err := s.pruningStore.PruningPointUTXOs(s.databaseContext, fromOutpoint, limit)
	if err != nil {
		return nil, err
	}
	return pruningPointUTXOs, nil
}

func (s *consensus) GetVirtualUTXOs(expectedVirtualParents []*externalapi.DomainHash,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	virtualParents, err := s.dagTopologyManager.Parents(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	if !externalapi.HashesEqual(expectedVirtualParents, virtualParents) {
		return nil, errors.Wrapf(ruleerrors.ErrGetVirtualUTXOsWrongVirtualParents, "expected virtual parents %s but got %s",
			expectedVirtualParents,
			virtualParents)
	}

	virtualUTXOs, err := s.consensusStateStore.VirtualUTXOs(s.databaseContext, fromOutpoint, limit)
	if err != nil {
		return nil, err
	}
	return virtualUTXOs, nil
}

func (s *consensus) PruningPoint() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningStore.PruningPoint(s.databaseContext)
}

func (s *consensus) ClearImportedPruningPointData() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.ClearImportedPruningPointData()
}

func (s *consensus) AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs)
}

func (s *consensus) ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainBlock) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertImportedPruningPoint(newPruningPoint)
}

func (s *consensus) GetVirtualSelectedParent() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	virtualGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	return virtualGHOSTDAGData.SelectedParent(), nil
}

func (s *consensus) Tips() ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.consensusStateStore.Tips(s.databaseContext)
}

func (s *consensus) GetVirtualInfo() (*externalapi.VirtualInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	blockRelations, err := s.blockRelationStore.BlockRelation(s.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	bits, err := s.difficultyManager.RequiredDifficulty(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	pastMedianTime, err := s.pastMedianTimeManager.PastMedianTime(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	virtualGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return &externalapi.VirtualInfo{
		ParentHashes:   blockRelations.Parents,
		Bits:           bits,
		PastMedianTime: pastMedianTime,
		BlueScore:      virtualGHOSTDAGData.BlueScore(),
	}, nil
}

func (s *consensus) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(lowHash)
	if err != nil {
		return nil, err
	}
	err = s.validateBlockHashExists(highHash)
	if err != nil {
		return nil, err
	}

	return s.syncManager.CreateBlockLocator(lowHash, highHash, limit)
}

func (s *consensus) CreateFullHeadersSelectedChainBlockLocator() (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	lowHash, err := s.pruningStore.PruningPoint(s.databaseContext)
	if err != nil {
		return nil, err
	}

	highHash, err := s.headersSelectedTipStore.HeadersSelectedTip(s.databaseContext)
	if err != nil {
		return nil, err
	}

	return s.syncManager.CreateHeadersSelectedChainBlockLocator(lowHash, highHash)
}

func (s *consensus) CreateHeadersSelectedChainBlockLocator(lowHash,
	highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.CreateHeadersSelectedChainBlockLocator(lowHash, highHash)
}

func (s *consensus) GetSyncInfo() (*externalapi.SyncInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.syncManager.GetSyncInfo()
}

func (s *consensus) IsValidPruningPoint(blockHash *externalapi.DomainHash) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(blockHash)
	if err != nil {
		return false, err
	}

	return s.pruningManager.IsValidPruningPoint(blockHash)
}

func (s *consensus) GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(blockHash)
	if err != nil {
		return nil, err
	}

	return s.consensusStateManager.GetVirtualSelectedParentChainFromBlock(blockHash)
}

func (s *consensus) validateBlockHashExists(blockHash *externalapi.DomainHash) error {
	exists, err := s.blockStatusStore.Exists(s.databaseContext, blockHash)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf("block %s does not exist", blockHash)
	}
	return nil
}

func (s *consensus) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(blockHashA)
	if err != nil {
		return false, err
	}
	err = s.validateBlockHashExists(blockHashB)
	if err != nil {
		return false, err
	}

	return s.dagTopologyManager.IsInSelectedParentChainOf(blockHashA, blockHashB)
}

func (s *consensus) GetHeadersSelectedTip() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.headersSelectedTipStore.HeadersSelectedTip(s.databaseContext)
}

func (s *consensus) Anticone(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(blockHash)
	if err != nil {
		return nil, err
	}

	return s.dagTraversalManager.Anticone(blockHash)
}
