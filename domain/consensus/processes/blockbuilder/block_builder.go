package blockbuilder

import (
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

	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	coinbaseManager       model.CoinbaseManager
	consensusStateManager model.ConsensusStateManager
	ghostdagManager       model.GHOSTDAGManager
	transactionValidator  model.TransactionValidator

	acceptanceDataStore model.AcceptanceDataStore
	blockRelationStore  model.BlockRelationStore
	multisetStore       model.MultisetStore
	ghostdagDataStore   model.GHOSTDAGDataStore
}

// New creates a new instance of a BlockBuilder
func New(
	databaseContext model.DBManager,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	coinbaseManager model.CoinbaseManager,
	consensusStateManager model.ConsensusStateManager,
	ghostdagManager model.GHOSTDAGManager,
	transactionValidator model.TransactionValidator,

	acceptanceDataStore model.AcceptanceDataStore,
	blockRelationStore model.BlockRelationStore,
	multisetStore model.MultisetStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
) model.BlockBuilder {

	return &blockBuilder{
		databaseContext:       databaseContext,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		coinbaseManager:       coinbaseManager,
		consensusStateManager: consensusStateManager,
		ghostdagManager:       ghostdagManager,
		transactionValidator:  transactionValidator,

		acceptanceDataStore: acceptanceDataStore,
		blockRelationStore:  blockRelationStore,
		multisetStore:       multisetStore,
		ghostdagDataStore:   ghostdagDataStore,
	}
}

// BuildBlock builds a block over the current state, with the given
// coinbaseData and the given transactions
func (bb *blockBuilder) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlock")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	return bb.buildBlock(stagingArea, coinbaseData, transactions)
}

func (bb *blockBuilder) buildBlock(stagingArea *model.StagingArea, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	err := bb.validateTransactions(stagingArea, transactions)
	if err != nil {
		return nil, err
	}

	coinbase, err := bb.newBlockCoinbaseTransaction(stagingArea, coinbaseData)
	if err != nil {
		return nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bb.buildHeader(stagingArea, transactionsWithCoinbase)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, nil
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

	err = bb.transactionValidator.ValidateTransactionInContextIgnoringUTXO(stagingArea, transaction, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	return bb.transactionValidator.ValidateTransactionInContextAndPopulateFee(stagingArea, transaction, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockCoinbaseTransaction(stagingArea *model.StagingArea,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error) {

	return bb.coinbaseManager.ExpectedCoinbaseTransaction(stagingArea, model.VirtualBlockHash, coinbaseData)
}

func (bb *blockBuilder) buildHeader(stagingArea *model.StagingArea, transactions []*externalapi.DomainTransaction) (
	externalapi.BlockHeader, error) {

	parentHashes, err := bb.newBlockParentHashes(stagingArea)
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

	return blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		parentHashes,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		timeInMilliseconds,
		bits,
		0,
	), nil
}

func (bb *blockBuilder) newBlockParentHashes(stagingArea *model.StagingArea) ([]*externalapi.DomainHash, error) {
	virtualBlockRelations, err := bb.blockRelationStore.BlockRelation(bb.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return virtualBlockRelations.Parents, nil
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
