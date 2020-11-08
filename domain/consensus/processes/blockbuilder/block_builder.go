package blockbuilder

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
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

	acceptanceDataStore model.AcceptanceDataStore
	blockRelationStore  model.BlockRelationStore
	multisetStore       model.MultisetStore
	ghostdagDataStore   model.GHOSTDAGDataStore
}

func New(
	databaseContext model.DBManager,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	coinbaseManager model.CoinbaseManager,
	consensusStateManager model.ConsensusStateManager,
	ghostdagManager model.GHOSTDAGManager,

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
		acceptanceDataStore:   acceptanceDataStore,
		blockRelationStore:    blockRelationStore,
		multisetStore:         multisetStore,
		ghostdagDataStore:     ghostdagDataStore,
	}
}

// BuildBlock builds a block over the current state, with the given
// coinbaseData and the given transactions
func (bb *blockBuilder) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlock")
	defer onEnd()

	return bb.buildBlock(coinbaseData, transactions)
}

func (bb *blockBuilder) buildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	coinbase, err := bb.newBlockCoinbaseTransaction(coinbaseData)
	if err != nil {
		return nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bb.buildHeader(transactionsWithCoinbase)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, nil
}

func (bb *blockBuilder) newBlockCoinbaseTransaction(
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error) {

	return bb.coinbaseManager.ExpectedCoinbaseTransaction(model.VirtualBlockHash, coinbaseData)
}

func (bb blockBuilder) buildHeader(transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlockHeader, error) {
	parentHashes, err := bb.newBlockParentHashes()
	if err != nil {
		return nil, err
	}
	virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	timeInMilliseconds, err := bb.newBlockTime(virtualGHOSTDAGData)
	if err != nil {
		return nil, err
	}
	bits, err := bb.newBlockDifficulty(virtualGHOSTDAGData)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)
	acceptedIDMerkleRoot, err := bb.newBlockAcceptedIDMerkleRoot()
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := bb.newBlockUTXOCommitment()
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlockHeader{
		Version:              constants.BlockVersion,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       *hashMerkleRoot,
		AcceptedIDMerkleRoot: *acceptedIDMerkleRoot,
		UTXOCommitment:       *utxoCommitment,
		TimeInMilliseconds:   timeInMilliseconds,
		Bits:                 bits,
	}, nil
}

func (bb blockBuilder) newBlockParentHashes() ([]*externalapi.DomainHash, error) {
	virtualBlockRelations, err := bb.blockRelationStore.BlockRelation(bb.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return virtualBlockRelations.Parents, nil
}

func (bb blockBuilder) newBlockTime(virtualGHOSTDAGData *model.BlockGHOSTDAGData) (int64, error) {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a millisecond boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// millisecond.
	newTimestamp := mstime.Now().UnixMilliseconds() + 1
	minTimestamp, err := bb.pastMedianTimeManager.PastMedianTime(virtualGHOSTDAGData.SelectedParent)
	if err != nil {
		return 0, err
	}
	if newTimestamp < minTimestamp {
		newTimestamp = minTimestamp
	}
	return newTimestamp, nil
}

func (bb blockBuilder) newBlockDifficulty(virtualGHOSTDAGData *model.BlockGHOSTDAGData) (uint32, error) {
	virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return 0, err
	}
	return bb.difficultyManager.RequiredDifficulty(virtualGHOSTDAGData.SelectedParent)
}

func (bb blockBuilder) newBlockHashMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	return merkle.CalculateHashMerkleRoot(transactions)
}

func (bb blockBuilder) newBlockAcceptedIDMerkleRoot() (*externalapi.DomainHash, error) {
	newBlockAcceptanceData, err := bb.acceptanceDataStore.Get(bb.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return bb.calculateAcceptedIDMerkleRoot(newBlockAcceptanceData)
}

func (bb blockBuilder) calculateAcceptedIDMerkleRoot(acceptanceData model.AcceptanceData) (*externalapi.DomainHash, error) {
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
		acceptedTransactionIID := consensusserialization.TransactionID(acceptedTransactions[i])
		acceptedTransactionJID := consensusserialization.TransactionID(acceptedTransactions[j])
		return transactionid.Less(acceptedTransactionIID, acceptedTransactionJID)
	})

	return merkle.CalculateIDMerkleRoot(acceptedTransactions), nil
}

func (bb blockBuilder) newBlockUTXOCommitment() (*externalapi.DomainHash, error) {
	newBlockMultiset, err := bb.multisetStore.Get(bb.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	newBlockUTXOCommitment := newBlockMultiset.Hash()
	return newBlockUTXOCommitment, nil
}
