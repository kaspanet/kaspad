package blockbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

type testBlockBuilder struct {
	*blockBuilder
	testConsensus testapi.TestConsensus
	nonceCounter  uint64
}

var tempBlockHash = externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})

// NewTestBlockBuilder creates an instance of a TestBlockBuilder
func NewTestBlockBuilder(baseBlockBuilder model.BlockBuilder, testConsensus testapi.TestConsensus) testapi.TestBlockBuilder {
	return &testBlockBuilder{
		blockBuilder:  baseBlockBuilder.(*blockBuilder),
		testConsensus: testConsensus,
	}
}

func cleanBlockPrefilledFields(block *externalapi.DomainBlock) {
	for _, tx := range block.Transactions {
		tx.Fee = 0
		tx.Mass = 0
		tx.ID = nil

		for _, input := range tx.Inputs {
			input.UTXOEntry = nil
		}
	}
}

// BuildBlockWithParents builds a block with provided parents, coinbaseData and transactions,
// and returns the block together with its past UTXO-diff from the virtual.
func (bb *testBlockBuilder) BuildBlockWithParents(parentHashes []*externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (
	*externalapi.DomainBlock, model.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlockWithParents")
	defer onEnd()

	block, diff, err := bb.buildBlockWithParents(parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, nil, err
	}

	// It's invalid to insert a block with prefilled fields to consensus, so we
	// clean them before returning the block.
	cleanBlockPrefilledFields(block)

	return block, diff, nil
}

func (bb *testBlockBuilder) buildUTXOInvalidHeader(parentHashes []*externalapi.DomainHash,
	transactions []*externalapi.DomainTransaction) (externalapi.BlockHeader, error) {

	timeInMilliseconds, err := bb.minBlockTime(tempBlockHash)
	if err != nil {
		return nil, err
	}

	bits, err := bb.difficultyManager.RequiredDifficulty(tempBlockHash)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)
	bb.nonceCounter++
	return blockheader.NewImmutableBlockHeader(
		constants.MaxBlockVersion,
		parentHashes,
		hashMerkleRoot,
		&externalapi.DomainHash{},
		&externalapi.DomainHash{},
		timeInMilliseconds,
		bits,
		bb.nonceCounter,
	), nil
}

func (bb *testBlockBuilder) buildHeaderWithParents(parentHashes []*externalapi.DomainHash,
	transactions []*externalapi.DomainTransaction, acceptanceData externalapi.AcceptanceData, multiset model.Multiset) (
	externalapi.BlockHeader, error) {

	header, err := bb.buildUTXOInvalidHeader(parentHashes, transactions)
	if err != nil {
		return nil, err
	}

	acceptedIDMerkleRoot, err := bb.calculateAcceptedIDMerkleRoot(acceptanceData)
	if err != nil {
		return nil, err
	}
	utxoCommitment := multiset.Hash()

	return blockheader.NewImmutableBlockHeader(
		header.Version(),
		header.ParentHashes(),
		header.HashMerkleRoot(),
		acceptedIDMerkleRoot,
		utxoCommitment,
		header.TimeInMilliseconds(),
		header.Bits(),
		header.Nonce(),
	), nil
}

func (bb *testBlockBuilder) buildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, model.UTXODiff, error) {

	defer bb.testConsensus.DiscardAllStores()

	if coinbaseData == nil {
		scriptPublicKey, _ := testutils.OpTrueScript()
		coinbaseData = &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
			ExtraData:       []byte{},
		}
	}

	bb.blockRelationStore.StageBlockRelation(tempBlockHash, &model.BlockRelations{Parents: parentHashes})

	err := bb.ghostdagManager.GHOSTDAG(tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	ghostdagData, err := bb.ghostdagDataStore.Get(bb.databaseContext, tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	selectedParentStatus, err := bb.testConsensus.ConsensusStateManager().ResolveBlockStatus(ghostdagData.SelectedParent())
	if err != nil {
		return nil, nil, err
	}
	if selectedParentStatus == externalapi.StatusDisqualifiedFromChain {
		return nil, nil, errors.Errorf("Error building block with selectedParent %s with status DisqualifiedFromChain",
			ghostdagData.SelectedParent())
	}

	pastUTXO, acceptanceData, multiset, err := bb.consensusStateManager.CalculatePastUTXOAndAcceptanceData(tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	bb.acceptanceDataStore.Stage(tempBlockHash, acceptanceData)

	coinbase, err := bb.coinbaseManager.ExpectedCoinbaseTransaction(tempBlockHash, coinbaseData)
	if err != nil {
		return nil, nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bb.buildHeaderWithParents(parentHashes, transactionsWithCoinbase, acceptanceData, multiset)
	if err != nil {
		return nil, nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, pastUTXO, nil
}

func (bb *testBlockBuilder) BuildUTXOInvalidHeader(parentHashes []*externalapi.DomainHash) (externalapi.BlockHeader,
	error) {

	block, err := bb.BuildUTXOInvalidBlock(parentHashes)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

func (bb *testBlockBuilder) BuildUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock,
	error) {

	defer bb.testConsensus.DiscardAllStores()

	bb.blockRelationStore.StageBlockRelation(tempBlockHash, &model.BlockRelations{Parents: parentHashes})

	err := bb.ghostdagManager.GHOSTDAG(tempBlockHash)
	if err != nil {
		return nil, err
	}

	// We use genesis transactions so we'll have something to build merkle root and coinbase with
	genesisTransactions := bb.testConsensus.DAGParams().GenesisBlock.Transactions
	header, err := bb.buildUTXOInvalidHeader(parentHashes, genesisTransactions)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: genesisTransactions,
	}, nil
}
