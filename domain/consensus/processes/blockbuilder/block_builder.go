package blockbuilder

import (
	"math/big"
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/mstime"
)

type blockBuilder struct {
	databaseContext model.DBManager
	genesisHash     *externalapi.DomainHash

	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	coinbaseManager       model.CoinbaseManager
	consensusStateManager model.ConsensusStateManager
	ghostdagManager       model.GHOSTDAGManager
	transactionValidator  model.TransactionValidator
	finalityManager       model.FinalityManager
	pruningManager        model.PruningManager
	blockParentBuilder    model.BlockParentBuilder

	acceptanceDataStore model.AcceptanceDataStore
	blockRelationStore  model.BlockRelationStore
	multisetStore       model.MultisetStore
	ghostdagDataStore   model.GHOSTDAGDataStore
	daaBlocksStore      model.DAABlocksStore
}

// New creates a new instance of a BlockBuilder
func New(
	databaseContext model.DBManager,
	genesisHash *externalapi.DomainHash,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	coinbaseManager model.CoinbaseManager,
	consensusStateManager model.ConsensusStateManager,
	ghostdagManager model.GHOSTDAGManager,
	transactionValidator model.TransactionValidator,
	finalityManager model.FinalityManager,
	blockParentBuilder model.BlockParentBuilder,
	pruningManager model.PruningManager,

	acceptanceDataStore model.AcceptanceDataStore,
	blockRelationStore model.BlockRelationStore,
	multisetStore model.MultisetStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	daaBlocksStore model.DAABlocksStore,
) model.BlockBuilder {

	return &blockBuilder{
		databaseContext: databaseContext,
		genesisHash:     genesisHash,

		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		coinbaseManager:       coinbaseManager,
		consensusStateManager: consensusStateManager,
		ghostdagManager:       ghostdagManager,
		transactionValidator:  transactionValidator,
		finalityManager:       finalityManager,
		blockParentBuilder:    blockParentBuilder,
		pruningManager:        pruningManager,

		acceptanceDataStore: acceptanceDataStore,
		blockRelationStore:  blockRelationStore,
		multisetStore:       multisetStore,
		ghostdagDataStore:   ghostdagDataStore,
		daaBlocksStore:      daaBlocksStore,
	}
}

// BuildBlock builds a block over the current state, with the given
// coinbaseData and the given transactions
func (bb *blockBuilder) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (block *externalapi.DomainBlock, coinbaseHasRedReward bool, err error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlock")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	return bb.buildBlock(stagingArea, coinbaseData, transactions)
}

func (bb *blockBuilder) buildBlock(stagingArea *model.StagingArea, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (block *externalapi.DomainBlock, coinbaseHasRedReward bool, err error) {

	err = bb.validateTransactions(stagingArea, transactions)
	if err != nil {
		return nil, false, err
	}

	newBlockPruningPoint, err := bb.newBlockPruningPoint(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, false, err
	}
	coinbase, coinbaseHasRedReward, err := bb.newBlockCoinbaseTransaction(stagingArea, coinbaseData)
	if err != nil {
		return nil, false, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bb.buildHeader(stagingArea, transactionsWithCoinbase, newBlockPruningPoint)
	if err != nil {
		return nil, false, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, coinbaseHasRedReward, nil
}

func (bb *blockBuilder) validateTransactions(stagingArea *model.StagingArea,
	transactions []*externalapi.DomainTransaction) error {

	invalidTransactions := make([]ruleerrors.InvalidTransaction, 0)
	for _, transaction := range transactions {
		err := bb.validateTransaction(stagingArea, transaction)
		if err != nil {
			if !errors.As(err, &ruleerrors.RuleError{}) {
				return err
			}
			invalidTransactions = append(invalidTransactions,
				ruleerrors.InvalidTransaction{Transaction: transaction, Error: err})
		}
	}

	if len(invalidTransactions) > 0 {
		return ruleerrors.NewErrInvalidTransactionsInNewBlock(invalidTransactions)
	}

	return nil
}

func (bb *blockBuilder) validateTransaction(
	stagingArea *model.StagingArea, transaction *externalapi.DomainTransaction) error {

	originalEntries := make([]externalapi.UTXOEntry, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		originalEntries[i] = input.UTXOEntry
		input.UTXOEntry = nil
	}

	defer func() {
		for i, input := range transaction.Inputs {
			input.UTXOEntry = originalEntries[i]
		}
	}()

	err := bb.consensusStateManager.PopulateTransactionWithUTXOEntries(stagingArea, transaction)
	if err != nil {
		return err
	}

	virtualPastMedianTime, err := bb.pastMedianTimeManager.PastMedianTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = bb.transactionValidator.ValidateTransactionInContextIgnoringUTXO(stagingArea, transaction, model.VirtualBlockHash, virtualPastMedianTime)
	if err != nil {
		return err
	}

	return bb.transactionValidator.ValidateTransactionInContextAndPopulateFee(stagingArea, transaction, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockCoinbaseTransaction(stagingArea *model.StagingArea,
	coinbaseData *externalapi.DomainCoinbaseData) (expectedTransaction *externalapi.DomainTransaction, hasRedReward bool, err error) {

	return bb.coinbaseManager.ExpectedCoinbaseTransaction(stagingArea, model.VirtualBlockHash, coinbaseData)
}

func (bb *blockBuilder) buildHeader(stagingArea *model.StagingArea, transactions []*externalapi.DomainTransaction,
	newBlockPruningPoint *externalapi.DomainHash) (externalapi.BlockHeader, error) {

	daaScore, err := bb.newBlockDAAScore(stagingArea)
	if err != nil {
		return nil, err
	}

	parents, err := bb.newBlockParents(stagingArea, daaScore)
	if err != nil {
		return nil, err
	}

	timeInMilliseconds, err := bb.newBlockTime(stagingArea)
	if err != nil {
		return nil, err
	}
	bits, err := bb.newBlockDifficulty(stagingArea)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)
	acceptedIDMerkleRoot, err := bb.newBlockAcceptedIDMerkleRoot(stagingArea)
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := bb.newBlockUTXOCommitment(stagingArea)
	if err != nil {
		return nil, err
	}
	blueWork, err := bb.newBlockBlueWork(stagingArea)
	if err != nil {
		return nil, err
	}
	blueScore, err := bb.newBlockBlueScore(stagingArea)
	if err != nil {
		return nil, err
	}

	return blockheader.NewImmutableBlockHeader(
		constants.BlockVersion,
		parents,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		timeInMilliseconds,
		bits,
		0,
		daaScore,
		blueScore,
		blueWork,
		newBlockPruningPoint,
	), nil
}

func (bb *blockBuilder) newBlockParents(stagingArea *model.StagingArea, daaScore uint64) ([]externalapi.BlockLevelParents, error) {
	virtualBlockRelations, err := bb.blockRelationStore.BlockRelation(bb.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	return bb.blockParentBuilder.BuildParents(stagingArea, daaScore, virtualBlockRelations.Parents)
}

func (bb *blockBuilder) newBlockTime(stagingArea *model.StagingArea) (int64, error) {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a millisecond boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// millisecond.
	newTimestamp := mstime.Now().UnixMilliseconds()
	minTimestamp, err := bb.minBlockTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return 0, err
	}
	if newTimestamp < minTimestamp {
		newTimestamp = minTimestamp
	}
	return newTimestamp, nil
}

func (bb *blockBuilder) minBlockTime(stagingArea *model.StagingArea, hash *externalapi.DomainHash) (int64, error) {
	pastMedianTime, err := bb.pastMedianTimeManager.PastMedianTime(stagingArea, hash)
	if err != nil {
		return 0, err
	}

	return pastMedianTime + 1, nil
}

func (bb *blockBuilder) newBlockDifficulty(stagingArea *model.StagingArea) (uint32, error) {
	return bb.difficultyManager.RequiredDifficulty(stagingArea, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockHashMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	return merkle.CalculateHashMerkleRoot(transactions)
}

func (bb *blockBuilder) newBlockAcceptedIDMerkleRoot(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	newBlockAcceptanceData, err := bb.acceptanceDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return bb.calculateAcceptedIDMerkleRoot(newBlockAcceptanceData)
}

func (bb *blockBuilder) calculateAcceptedIDMerkleRoot(acceptanceData externalapi.AcceptanceData) (*externalapi.DomainHash, error) {
	var acceptedTransactions []*externalapi.DomainTransaction
	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptance := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptance.IsAccepted {
				continue
			}
			acceptedTransactions = append(acceptedTransactions, transactionAcceptance.Transaction)
		}
	}
	sort.Slice(acceptedTransactions, func(i, j int) bool {
		acceptedTransactionIID := consensushashing.TransactionID(acceptedTransactions[i])
		acceptedTransactionJID := consensushashing.TransactionID(acceptedTransactions[j])
		return acceptedTransactionIID.Less(acceptedTransactionJID)
	})

	return merkle.CalculateIDMerkleRoot(acceptedTransactions), nil
}

func (bb *blockBuilder) newBlockUTXOCommitment(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	newBlockMultiset, err := bb.multisetStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	newBlockUTXOCommitment := newBlockMultiset.Hash()
	return newBlockUTXOCommitment, nil
}

func (bb *blockBuilder) newBlockDAAScore(stagingArea *model.StagingArea) (uint64, error) {
	return bb.daaBlocksStore.DAAScore(bb.databaseContext, stagingArea, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockBlueWork(stagingArea *model.StagingArea) (*big.Int, error) {
	virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}
	return virtualGHOSTDAGData.BlueWork(), nil
}

func (bb *blockBuilder) newBlockBlueScore(stagingArea *model.StagingArea) (uint64, error) {
	virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return 0, err
	}
	return virtualGHOSTDAGData.BlueScore(), nil
}

func (bb *blockBuilder) newBlockPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	return bb.pruningManager.ExpectedHeaderPruningPoint(stagingArea, blockHash)
}
