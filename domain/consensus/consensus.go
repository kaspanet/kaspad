package consensus

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"math/big"
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

	genesisBlock *externalapi.DomainBlock
	genesisHash  *externalapi.DomainHash

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
	daaBlocksStore            model.DAABlocksStore
}

func (s *consensus) ValidateAndInsertBlockWithMetaData(block *externalapi.BlockWithMetaData, validateUTXO bool) (*externalapi.BlockInsertionResult, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertBlockWithMetaData(block, validateUTXO)
}

// Init initializes consensus
func (s *consensus) Init(shouldNotAddGenesis bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	onEnd := logger.LogAndMeasureExecutionTime(log, "Init")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	exists, err := s.blockStatusStore.Exists(s.databaseContext, stagingArea, model.VirtualGenesisBlockHash)
	if err != nil {
		return err
	}

	if !exists {
		s.blockStatusStore.Stage(stagingArea, model.VirtualGenesisBlockHash, externalapi.StatusUTXOValid)
		err = s.reachabilityManager.Init(stagingArea)
		if err != nil {
			return err
		}

		err = s.dagTopologyManager.SetParents(stagingArea, model.VirtualGenesisBlockHash, nil)
		if err != nil {
			return err
		}

		s.consensusStateStore.StageTips(stagingArea, []*externalapi.DomainHash{model.VirtualGenesisBlockHash})
		s.ghostdagDataStore.Stage(stagingArea, model.VirtualGenesisBlockHash, externalapi.NewBlockGHOSTDAGData(
			0,
			big.NewInt(0),
			nil,
			nil,
			nil,
			nil,
		))

		err = staging.CommitAllChanges(s.databaseContext, stagingArea)
		if err != nil {
			return err
		}
	}

	if !shouldNotAddGenesis && s.blockStore.Count(stagingArea) == 0 {
		genesisWithMetaData := &externalapi.BlockWithMetaData{
			Block:     s.genesisBlock,
			DAAScore:  0,
			DAAWindow: nil,
			GHOSTDAGData: []*externalapi.BlockGHOSTDAGDataHashPair{
				{
					GHOSTDAGData: externalapi.NewBlockGHOSTDAGData(0, big.NewInt(0), model.VirtualGenesisBlockHash, nil, nil, make(map[externalapi.DomainHash]externalapi.KType)),
					Hash:         s.genesisHash,
				},
			},
		}
		_, err = s.blockProcessor.ValidateAndInsertBlockWithMetaData(genesisWithMetaData, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *consensus) ResolveVirtual() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	err := s.consensusStateManager.ResolveVirtual(stagingArea)
	if err != nil {
		return err
	}

	return staging.CommitAllChanges(s.databaseContext, stagingArea)
}

func (s *consensus) PruningPointAndItsAnticoneWithMetaData() ([]*externalapi.BlockWithMetaData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.PruningPointAndItsAnticoneWithMetaData()
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
func (s *consensus) ValidateAndInsertBlock(block *externalapi.DomainBlock, validateUTXO bool) (*externalapi.BlockInsertionResult, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertBlock(block, validateUTXO)
}

// ValidateTransactionAndPopulateWithConsensusData validates the given transaction
// and populates it with any missing consensus data
func (s *consensus) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.transactionValidator.ValidateTransactionInIsolation(transaction)
	if err != nil {
		return err
	}

	err = s.consensusStateManager.PopulateTransactionWithUTXOEntries(stagingArea, transaction)
	if err != nil {
		return err
	}

	virtualSelectedParentMedianTime, err := s.pastMedianTimeManager.PastMedianTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	return s.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(
		stagingArea, transaction, model.VirtualBlockHash, virtualSelectedParentMedianTime)
}

func (s *consensus) GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	block, err := s.blockStore.Block(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, errors.Wrapf(err, "block %s does not exist", blockHash)
		}
		return nil, err
	}
	return block, nil
}

func (s *consensus) GetBlockEvenIfHeaderOnly(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	block, err := s.blockStore.Block(s.databaseContext, stagingArea, blockHash)
	if err == nil {
		return block, nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	header, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, errors.Wrapf(err, "block %s does not exist", blockHash)
		}
		return nil, err
	}
	return &externalapi.DomainBlock{Header: header}, nil
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockHeader, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, blockHash)
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

	stagingArea := model.NewStagingArea()

	blockInfo := &externalapi.BlockInfo{}

	exists, err := s.blockStatusStore.Exists(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.Exists = exists
	if !exists {
		return blockInfo, nil
	}

	blockStatus, err := s.blockStatusStore.Get(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.BlockStatus = blockStatus

	// If the status is invalid, then we don't have the necessary reachability data to check if it's in PruningPoint.Future.
	if blockStatus == externalapi.StatusInvalid {
		return blockInfo, nil
	}

	ghostdagData, err := s.ghostdagDataStore.Get(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	blockInfo.BlueScore = ghostdagData.BlueScore()
	blockInfo.BlueWork = ghostdagData.BlueWork()

	return blockInfo, nil
}

func (s *consensus) GetBlockRelations(blockHash *externalapi.DomainHash) (
	parents []*externalapi.DomainHash, selectedParent *externalapi.DomainHash,
	children []*externalapi.DomainHash, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockRelation, err := s.blockRelationStore.BlockRelation(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	blockGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	return blockRelation.Parents, blockGHOSTDAGData.SelectedParent(), blockRelation.Children, nil
}

func (s *consensus) GetBlockAcceptanceData(blockHash *externalapi.DomainHash) (externalapi.AcceptanceData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return s.acceptanceDataStore.Get(s.databaseContext, stagingArea, blockHash)
}

func (s *consensus) GetHashesBetween(lowHash, highHash *externalapi.DomainHash, maxBlocks uint64) (
	hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err = s.validateBlockHashExists(stagingArea, lowHash)
	if err != nil {
		return nil, nil, err
	}
	err = s.validateBlockHashExists(stagingArea, highHash)
	if err != nil {
		return nil, nil, err
	}

	return s.syncManager.GetHashesBetween(stagingArea, lowHash, highHash, maxBlocks)
}

func (s *consensus) GetPruningPointUTXOs(expectedPruningPointHash *externalapi.DomainHash,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	pruningPointHash, err := s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
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

	stagingArea := model.NewStagingArea()

	virtualParents, err := s.dagTopologyManager.Parents(stagingArea, model.VirtualBlockHash)
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

	stagingArea := model.NewStagingArea()

	return s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
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

func (s *consensus) ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainHash) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertImportedPruningPoint(newPruningPoint)
}

func (s *consensus) GetVirtualSelectedParent() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	virtualGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	return virtualGHOSTDAGData.SelectedParent(), nil
}

func (s *consensus) Tips() ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.consensusStateStore.Tips(stagingArea, s.databaseContext)
}

func (s *consensus) GetVirtualInfo() (*externalapi.VirtualInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockRelations, err := s.blockRelationStore.BlockRelation(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	bits, err := s.difficultyManager.RequiredDifficulty(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	pastMedianTime, err := s.pastMedianTimeManager.PastMedianTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	virtualGHOSTDAGData, err := s.ghostdagDataStore.Get(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	daaScore, err := s.daaBlocksStore.DAAScore(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return &externalapi.VirtualInfo{
		ParentHashes:   blockRelations.Parents,
		Bits:           bits,
		PastMedianTime: pastMedianTime,
		BlueScore:      virtualGHOSTDAGData.BlueScore(),
		DAAScore:       daaScore,
	}, nil
}

func (s *consensus) GetVirtualDAAScore() (uint64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.daaBlocksStore.DAAScore(s.databaseContext, stagingArea, model.VirtualBlockHash)
}

func (s *consensus) CreateBlockLocator(lowHash, highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	if lowHash == nil {
		pruningPoint, err := s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
		if err != nil {
			return nil, err
		}

		lowHash = pruningPoint
	}

	err := s.validateBlockHashExists(stagingArea, lowHash)
	if err != nil {
		return nil, err
	}
	err = s.validateBlockHashExists(stagingArea, highHash)
	if err != nil {
		return nil, err
	}

	return s.syncManager.CreateBlockLocator(stagingArea, lowHash, highHash, limit)
}

func (s *consensus) CreateFullHeadersSelectedChainBlockLocator() (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	lowHash, err := s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	highHash, err := s.headersSelectedTipStore.HeadersSelectedTip(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	return s.syncManager.CreateHeadersSelectedChainBlockLocator(stagingArea, lowHash, highHash)
}

func (s *consensus) CreateHeadersSelectedChainBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.syncManager.CreateHeadersSelectedChainBlockLocator(stagingArea, lowHash, highHash)
}

func (s *consensus) GetSyncInfo() (*externalapi.SyncInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.syncManager.GetSyncInfo(stagingArea)
}

func (s *consensus) IsValidPruningPoint(blockHash *externalapi.DomainHash) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	return s.pruningManager.IsValidPruningPoint(stagingArea, blockHash)
}

func (s *consensus) GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return s.consensusStateManager.GetVirtualSelectedParentChainFromBlock(stagingArea, blockHash)
}

func (s *consensus) validateBlockHashExists(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	exists, err := s.blockStatusStore.Exists(s.databaseContext, stagingArea, blockHash)
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

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHashA)
	if err != nil {
		return false, err
	}
	err = s.validateBlockHashExists(stagingArea, blockHashB)
	if err != nil {
		return false, err
	}

	return s.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, blockHashA, blockHashB)
}

func (s *consensus) GetHeadersSelectedTip() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.headersSelectedTipStore.HeadersSelectedTip(s.databaseContext, stagingArea)
}

func (s *consensus) Anticone(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return s.dagTraversalManager.Anticone(stagingArea, blockHash)
}

func (s *consensus) EstimateNetworkHashesPerSecond(startHash *externalapi.DomainHash, windowSize int) (uint64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.difficultyManager.EstimateNetworkHashesPerSecond(startHash, windowSize)
}
