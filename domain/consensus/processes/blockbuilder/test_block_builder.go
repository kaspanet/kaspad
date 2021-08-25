package blockbuilder

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
	"math/big"
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
	*externalapi.DomainBlock, externalapi.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlockWithParents")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	block, diff, err := bb.buildBlockWithParents(stagingArea, parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, nil, err
	}

	// It's invalid to insert a block with prefilled fields to consensus, so we
	// clean them before returning the block.
	cleanBlockPrefilledFields(block)

	return block, diff, nil
}

func (bb *testBlockBuilder) buildUTXOInvalidHeader(stagingArea *model.StagingArea,
	parentHashes []*externalapi.DomainHash, bits uint32, daaScore, blueScore uint64, blueWork *big.Int,
	transactions []*externalapi.DomainTransaction) (externalapi.BlockHeader, error) {

	timeInMilliseconds, err := bb.minBlockTime(stagingArea, tempBlockHash)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)

	pruningPoint, err := bb.newBlockPruningPoint(stagingArea, tempBlockHash)
	if err != nil {
		return nil, err
	}

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
		daaScore,
		blueScore,
		blueWork,
		pruningPoint,
	), nil
}

func (bb *testBlockBuilder) buildHeaderWithParents(stagingArea *model.StagingArea,
	parentHashes []*externalapi.DomainHash, bits uint32, transactions []*externalapi.DomainTransaction,
	acceptanceData externalapi.AcceptanceData, multiset model.Multiset, daaScore, blueScore uint64, blueWork *big.Int) (externalapi.BlockHeader, error) {

	header, err := bb.buildUTXOInvalidHeader(stagingArea, parentHashes, bits, daaScore, blueScore, blueWork, transactions)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)
	acceptedIDMerkleRoot, err := bb.calculateAcceptedIDMerkleRoot(acceptanceData)
	if err != nil {
		return nil, err
	}
	utxoCommitment := multiset.Hash()

	return blockheader.NewImmutableBlockHeader(
		header.Version(),
		header.ParentHashes(),
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		header.TimeInMilliseconds(),
		header.Bits(),
		header.Nonce(),
		header.DAAScore(),
		header.BlueScore(),
		header.BlueWork(),
		header.PruningPoint(),
	), nil
}

func (bb *testBlockBuilder) buildBlockWithParents(stagingArea *model.StagingArea, parentHashes []*externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (
	*externalapi.DomainBlock, externalapi.UTXODiff, error) {

	if coinbaseData == nil {
		scriptPublicKeyScript, err := txscript.PayToScriptHashScript([]byte{txscript.OpTrue})
		if err != nil {
			panic(errors.Wrapf(err, "Couldn't parse opTrueScript. This should never happen"))
		}
		scriptPublicKey := &externalapi.ScriptPublicKey{Script: scriptPublicKeyScript, Version: constants.MaxScriptPublicKeyVersion}
		coinbaseData = &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
			ExtraData:       []byte{},
		}
	}

	bb.blockRelationStore.StageBlockRelation(stagingArea, tempBlockHash, &model.BlockRelations{Parents: parentHashes})

	err := bb.ghostdagManager.GHOSTDAG(stagingArea, tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	bits, err := bb.difficultyManager.StageDAADataAndReturnRequiredDifficulty(stagingArea, tempBlockHash, false)
	if err != nil {
		return nil, nil, err
	}
	daaScore, err := bb.daaBlocksStore.DAAScore(bb.databaseContext, stagingArea, tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	ghostdagData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, tempBlockHash, false)
	if err != nil {
		return nil, nil, err
	}
	blueWork := ghostdagData.BlueWork()
	blueScore := ghostdagData.BlueScore()

	selectedParentStatus, err := bb.testConsensus.ConsensusStateManager().ResolveBlockStatus(
		stagingArea, ghostdagData.SelectedParent(), false)
	if err != nil {
		return nil, nil, err
	}
	if selectedParentStatus == externalapi.StatusDisqualifiedFromChain {
		return nil, nil, errors.Errorf("Error building block with selectedParent %s with status DisqualifiedFromChain",
			ghostdagData.SelectedParent())
	}

	pastUTXO, acceptanceData, multiset, err :=
		bb.consensusStateManager.CalculatePastUTXOAndAcceptanceData(stagingArea, tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	bb.acceptanceDataStore.Stage(stagingArea, tempBlockHash, acceptanceData)

	coinbase, err := bb.coinbaseManager.ExpectedCoinbaseTransaction(stagingArea, tempBlockHash, coinbaseData)
	if err != nil {
		return nil, nil, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	err = bb.testConsensus.ReachabilityManager().AddBlock(stagingArea, tempBlockHash)
	if err != nil {
		return nil, nil, err
	}

	header, err := bb.buildHeaderWithParents(
		stagingArea, parentHashes, bits, transactionsWithCoinbase, acceptanceData, multiset, daaScore, blueScore, blueWork)
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

	stagingArea := model.NewStagingArea()

	bb.blockRelationStore.StageBlockRelation(stagingArea, tempBlockHash, &model.BlockRelations{Parents: parentHashes})

	err := bb.ghostdagManager.GHOSTDAG(stagingArea, tempBlockHash)
	if err != nil {
		return nil, err
	}

	bits, err := bb.difficultyManager.StageDAADataAndReturnRequiredDifficulty(stagingArea, tempBlockHash, false)
	if err != nil {
		return nil, err
	}
	daaScore, err := bb.daaBlocksStore.DAAScore(bb.databaseContext, stagingArea, tempBlockHash)
	if err != nil {
		return nil, err
	}

	ghostdagData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, tempBlockHash, false)
	if err != nil {
		return nil, err
	}
	blueWork := ghostdagData.BlueWork()
	blueScore := ghostdagData.BlueScore()

	// We use the genesis coinbase so that we'll have something to build merkle root and a new coinbase with
	genesisTransactions := bb.testConsensus.DAGParams().GenesisBlock.Transactions
	genesisCoinbase := genesisTransactions[transactionhelper.CoinbaseTransactionIndex].Clone()
	binary.LittleEndian.PutUint64(genesisCoinbase.Payload[:8], ghostdagData.BlueScore())
	transactions := []*externalapi.DomainTransaction{genesisCoinbase}

	err = bb.testConsensus.ReachabilityManager().AddBlock(stagingArea, tempBlockHash)
	if err != nil {
		return nil, err
	}

	header, err := bb.buildUTXOInvalidHeader(stagingArea, parentHashes, bits, daaScore, blueScore, blueWork, transactions)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
	}, nil
}
