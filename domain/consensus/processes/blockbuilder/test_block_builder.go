package blockbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
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

var tempBlockHash = &externalapi.DomainHash{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// NewTestBlockBuilder creates an instance of a TestBlockBuilder
func NewTestBlockBuilder(baseBlockBuilder model.BlockBuilder, testConsensus testapi.TestConsensus) model.TestBlockBuilder {
	return &testBlockBuilder{
		blockBuilder:  baseBlockBuilder.(*blockBuilder),
		testConsensus: testConsensus,
	}
}

func (bb *testBlockBuilder) BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlockWithParents")
	defer onEnd()

	return bb.buildBlockWithParents(parentHashes, coinbaseData, transactions)
}

func (bb *testBlockBuilder) buildHeaderWithParents(parentHashes []*externalapi.DomainHash,
	transactions []*externalapi.DomainTransaction, acceptanceData model.AcceptanceData, multiset model.Multiset) (
	*externalapi.DomainBlockHeader, error) {

	timeInMilliseconds, err := bb.minBlockTime(tempBlockHash)
	if err != nil {
		return nil, err
	}

	bits, err := bb.difficultyManager.RequiredDifficulty(tempBlockHash)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)
	acceptedIDMerkleRoot, err := bb.calculateAcceptedIDMerkleRoot(acceptanceData)
	if err != nil {
		return nil, err
	}
	utxoCommitment := multiset.Hash()

	bb.nonceCounter++
	return &externalapi.DomainBlockHeader{
		Version:              constants.BlockVersion,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       *hashMerkleRoot,
		AcceptedIDMerkleRoot: *acceptedIDMerkleRoot,
		UTXOCommitment:       *utxoCommitment,
		TimeInMilliseconds:   timeInMilliseconds,
		Bits:                 bits,
		Nonce:                bb.nonceCounter,
	}, nil
}

func (bb *testBlockBuilder) buildBlockWithParents(
	parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

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
		return nil, err
	}

	ghostdagData, err := bb.ghostdagDataStore.Get(bb.databaseContext, tempBlockHash)
	if err != nil {
		return nil, err
	}

	selectedParentStatus, err := bb.testConsensus.ConsensusStateManager().ResolveBlockStatus(ghostdagData.SelectedParent)
	if err != nil {
		return nil, err
	}
	if selectedParentStatus == externalapi.StatusDisqualifiedFromChain {
		return nil, errors.Errorf("Error building block with selectedParent %s with status DisqualifiedFromChain",
			ghostdagData.SelectedParent)
	}

	_, acceptanceData, multiset, err := bb.consensusStateManager.CalculatePastUTXOAndAcceptanceData(tempBlockHash)
	if err != nil {
		return nil, err
	}

	bb.acceptanceDataStore.Stage(tempBlockHash, acceptanceData)

	coinbase, err := bb.coinbaseManager.ExpectedCoinbaseTransaction(tempBlockHash, coinbaseData)
	if err != nil {
		return nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bb.buildHeaderWithParents(parentHashes, transactionsWithCoinbase, acceptanceData, multiset)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, nil
}
