package blockbuilder

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/infrastructure/logger"
)

type testBlockBuilder struct {
	*blockBuilder
}

var tempBlockHash = &externalapi.DomainHash{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// NewTestBlockBuilder creates an instance of a TestBlockBuilder
func NewTestBlockBuilder(baseBlockBuilder model.BlockBuilder) model.TestBlockBuilder {
	return &testBlockBuilder{blockBuilder: baseBlockBuilder.(*blockBuilder)}
}

func (bb *testBlockBuilder) BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlockWithParents")
	defer onEnd()

	return bb.buildBlockWithParents(parentHashes, coinbaseData, transactions)
}

func (bb testBlockBuilder) buildHeaderWithParents(parentHashes []*externalapi.DomainHash,
	transactions []*externalapi.DomainTransaction, acceptanceData model.AcceptanceData, multiset model.Multiset) (
	*externalapi.DomainBlockHeader, error) {

	ghostdagData, err := bb.ghostdagDataStore.Get(bb.databaseContext, tempBlockHash)
	if err != nil {
		return nil, err
	}
	timeInMilliseconds, err := bb.newBlockTime()
	if err != nil {
		return nil, err
	}
	bits, err := bb.newBlockDifficulty(ghostdagData)
	if err != nil {
		return nil, err
	}
	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)
	acceptedIDMerkleRoot, err := bb.calculateAcceptedIDMerkleRoot(acceptanceData)
	if err != nil {
		return nil, err
	}
	utxoCommitment := multiset.Hash()

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

func (bb *testBlockBuilder) buildBlockWithParents(
	parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	err := bb.blockRelationStore.StageBlockRelation(tempBlockHash, &model.BlockRelations{Parents: parentHashes})
	if err != nil {
		return nil, err
	}
	defer bb.blockRelationStore.Discard()

	err = bb.ghostdagManager.GHOSTDAG(tempBlockHash)
	if err != nil {
		return nil, err
	}
	defer bb.ghostdagDataStore.Discard()

	_, acceptanceData, multiset, err := bb.consensusStateManager.CalculatePastUTXOAndAcceptanceData(tempBlockHash)
	if err != nil {
		return nil, err
	}
	err = bb.acceptanceDataStore.Stage(tempBlockHash, acceptanceData)
	if err != nil {
		return nil, err
	}
	defer bb.acceptanceDataStore.Discard()

	coinbase, err := bb.newBlockCoinbaseTransaction(coinbaseData)
	if err != nil {
		return nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	header, err := bb.buildHeaderWithParents(parentHashes, transactions, acceptanceData, multiset)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, nil
}
