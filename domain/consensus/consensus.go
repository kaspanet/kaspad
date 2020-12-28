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

	acceptanceDataStore     model.AcceptanceDataStore
	blockStore              model.BlockStore
	blockHeaderStore        model.BlockHeaderStore
	pruningStore            model.PruningStore
	ghostdagDataStore       model.GHOSTDAGDataStore
	blockRelationStore      model.BlockRelationStore
	blockStatusStore        model.BlockStatusStore
	consensusStateStore     model.ConsensusStateStore
	headersSelectedTipStore model.HeaderSelectedTipStore
	multisetStore           model.MultisetStore
	reachabilityDataStore   model.ReachabilityDataStore
	utxoDiffStore           model.UTXODiffStore
	finalityStore           model.FinalityStore
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

	err := s.validateBlockHashExists(blockHash)
	if err != nil {
		return nil, err
	}

	return s.blockStore.Block(s.databaseContext, blockHash)
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (externalapi.ImmutableBlockHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.validateBlockHashExists(blockHash)
	if err != nil {
		return nil, err
	}

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

func (s *consensus) GetPruningPointUTXOSet(expectedPruningPointHash *externalapi.DomainHash) ([]byte, error) {
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

	serializedUTXOSet, err := s.pruningStore.PruningPointSerializedUTXOSet(s.databaseContext)
	if err != nil {
		return nil, err
	}
	return serializedUTXOSet, nil
}

func (s *consensus) PruningPoint() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningStore.PruningPoint(s.databaseContext)
}

func (s *consensus) ValidateAndInsertPruningPoint(newPruningPoint *externalapi.DomainBlock, serializedUTXOSet []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertPruningPoint(newPruningPoint, serializedUTXOSet)
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

func (s *consensus) FindNextBlockLocatorBoundaries(blockLocator externalapi.BlockLocator) (lowHash, highHash *externalapi.DomainHash, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(blockLocator) == 0 {
		return nil, nil, errors.Errorf("empty block locator")
	}

	return s.syncManager.FindNextBlockLocatorBoundaries(blockLocator)
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

func (s *consensus) GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error) {
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
		return errors.Errorf("block %s does not exists", blockHash)
	}
	return nil
}
