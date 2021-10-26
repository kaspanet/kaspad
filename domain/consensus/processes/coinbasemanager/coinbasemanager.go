package coinbasemanager

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
)

type coinbaseManager struct {
	subsidyGenesisReward                    uint64
	minSubsidy                              uint64
	maxSubsidy                              uint64
	subsidyPastRewardMultiplier             *big.Rat
	subsidyMergeSetRewardMultiplier         *big.Rat
	coinbasePayloadScriptPublicKeyMaxLength uint8
	genesisHash                             *externalapi.DomainHash

	databaseContext     model.DBReader
	dagTraversalManager model.DAGTraversalManager
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
	daaBlocksStore      model.DAABlocksStore
	blockStore          model.BlockStore
}

func (c *coinbaseManager) ExpectedCoinbaseTransaction(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error) {

	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	acceptanceData, err := c.acceptanceDataStore.Get(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	daaAddedBlocksSet, err := c.daaAddedBlocksSet(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	txOuts := make([]*externalapi.DomainTransactionOutput, 0, len(ghostdagData.MergeSetBlues()))
	totalSubsidy := uint64(0)
	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	for _, blue := range ghostdagData.MergeSetBlues() {
		txOut, subsidy, hasReward, err := c.coinbaseOutputAndSubsidyForBlueBlock(stagingArea, blue, acceptanceDataMap[*blue], daaAddedBlocksSet)
		if err != nil {
			return nil, err
		}

		if hasReward {
			totalSubsidy += subsidy
			txOuts = append(txOuts, txOut)
		}
	}

	txOut, subsidy, hasReward, err := c.coinbaseOutputAndSubsidyForRewardFromRedBlocks(
		stagingArea, ghostdagData, acceptanceData, daaAddedBlocksSet, coinbaseData)
	if err != nil {
		return nil, err
	}

	if hasReward {
		totalSubsidy += subsidy
		txOuts = append(txOuts, txOut)
	}

	payload, err := c.serializeCoinbasePayload(ghostdagData.BlueScore(), coinbaseData, totalSubsidy)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      txOuts,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          0,
		Payload:      payload,
	}, nil
}

func (c *coinbaseManager) daaAddedBlocksSet(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	hashset.HashSet, error) {

	daaAddedBlocks, err := c.daaBlocksStore.DAAAddedBlocks(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return hashset.NewFromSlice(daaAddedBlocks...), nil
}

// coinbaseOutputAndSubsidyForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func (c *coinbaseManager) coinbaseOutputAndSubsidyForBlueBlock(stagingArea *model.StagingArea,
	blueBlock *externalapi.DomainHash, blockAcceptanceData *externalapi.BlockAcceptanceData,
	mergingBlockDAAAddedBlocksSet hashset.HashSet) (*externalapi.DomainTransactionOutput, uint64, bool, error) {

	subsidy, totalFees, err := c.calcMergedBlockReward(stagingArea, blueBlock, blockAcceptanceData, mergingBlockDAAAddedBlocksSet)
	if err != nil {
		return nil, 0, false, err
	}

	totalReward := subsidy + totalFees
	if totalReward == 0 {
		return nil, 0, false, nil
	}

	// the ScriptPublicKey for the coinbase is parsed from the coinbase payload
	_, coinbaseData, _, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(blockAcceptanceData.TransactionAcceptanceData[0].Transaction)
	if err != nil {
		return nil, 0, false, err
	}

	txOut := &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}

	return txOut, subsidy, true, nil
}

func (c *coinbaseManager) coinbaseOutputAndSubsidyForRewardFromRedBlocks(stagingArea *model.StagingArea,
	ghostdagData *externalapi.BlockGHOSTDAGData, acceptanceData externalapi.AcceptanceData, daaAddedBlocksSet hashset.HashSet,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransactionOutput, uint64, bool, error) {

	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	totalSubsidy := uint64(0)
	totalReward := uint64(0)
	for _, red := range ghostdagData.MergeSetReds() {
		subsidy, totalFees, err := c.calcMergedBlockReward(stagingArea, red, acceptanceDataMap[*red], daaAddedBlocksSet)
		if err != nil {
			return nil, 0, false, err
		}

		totalSubsidy += subsidy

		reward := subsidy + totalFees
		totalReward += reward
	}

	if totalReward == 0 {
		return nil, 0, false, nil
	}

	return &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}, totalSubsidy, true, nil
}

func acceptanceDataFromArrayToMap(acceptanceData externalapi.AcceptanceData) map[externalapi.DomainHash]*externalapi.BlockAcceptanceData {
	acceptanceDataMap := make(map[externalapi.DomainHash]*externalapi.BlockAcceptanceData, len(acceptanceData))
	for _, blockAcceptanceData := range acceptanceData {
		acceptanceDataMap[*blockAcceptanceData.BlockHash] = blockAcceptanceData
	}
	return acceptanceDataMap
}

// calcBlockSubsidy returns the subsidy amount a block at the provided blue score
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// Further details: https://hashdag.medium.com/kaspa-launch-plan-9a63f4d754a6
func (c *coinbaseManager) calcBlockSubsidy(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	if blockHash.Equal(c.genesisHash) {
		return c.subsidyGenesisReward, nil
	}

	averagePastSubsidy, err := c.calculateAveragePastSubsidy(stagingArea, blockHash)
	if err != nil {
		return 0, err
	}
	mergeSetSubsidySum, err := c.calculateMergeSetSubsidySum(stagingArea, blockHash)
	if err != nil {
		return 0, err
	}
	subsidyRandomVariable, err := c.calculateSubsidyRandomVariable(stagingArea, blockHash)
	if err != nil {
		return 0, err
	}

	pastSubsidy := new(big.Rat).Mul(averagePastSubsidy, c.subsidyPastRewardMultiplier)
	mergeSetSubsidy := new(big.Rat).Mul(mergeSetSubsidySum, c.subsidyMergeSetRewardMultiplier)

	// In order to avoid unsupported negative exponents in powInt64, flip
	// the numerator and the denominator manually
	subsidyRandom := new(big.Rat)
	if subsidyRandomVariable >= 0 {
		subsidyRandom = subsidyRandom.SetInt64(powInt64(2, subsidyRandomVariable))
	} else {
		subsidyRandom = subsidyRandom.SetFrac64(1, powInt64(2, -subsidyRandomVariable))
	}

	blockSubsidyBigRat := new(big.Rat).Add(mergeSetSubsidy, new(big.Rat).Mul(pastSubsidy, subsidyRandom))
	blockSubsidyBigInt := new(big.Int).Div(blockSubsidyBigRat.Num(), blockSubsidyBigRat.Denom())
	blockSubsidyUint64 := blockSubsidyBigInt.Uint64()

	clampedBlockSubsidy := blockSubsidyUint64
	if clampedBlockSubsidy < c.minSubsidy {
		clampedBlockSubsidy = c.minSubsidy
	} else if clampedBlockSubsidy > c.maxSubsidy {
		clampedBlockSubsidy = c.maxSubsidy
	}

	return clampedBlockSubsidy, nil
}

func (c *coinbaseManager) calculateAveragePastSubsidy(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*big.Rat, error) {
	const subsidyPastWindowSize = 100
	blockWindow, err := c.dagTraversalManager.BlockWindow(stagingArea, blockHash, subsidyPastWindowSize)
	if err != nil {
		return nil, err
	}
	if len(blockWindow) == 0 {
		return new(big.Rat), nil
	}

	pastBlocks, err := c.blockStore.Blocks(c.databaseContext, stagingArea, blockWindow)
	if err != nil {
		return nil, err
	}

	pastBlockSubsidySum := int64(0)
	for _, pastBlock := range pastBlocks {
		coinbaseTransaction := pastBlock.Transactions[transactionhelper.CoinbaseTransactionIndex]
		_, _, pastBlockSubsidy, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(coinbaseTransaction)
		if err != nil {
			return nil, err
		}
		pastBlockSubsidySum += int64(pastBlockSubsidy)
	}
	return big.NewRat(pastBlockSubsidySum, int64(len(blockWindow))), nil
}

func (c *coinbaseManager) calculateMergeSetSubsidySum(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*big.Rat, error) {
	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	// Merge set blues containing nothing but the virtual genesis indicates that
	// this block is trusted. Get the data again with isTrustedData = true
	if len(ghostdagData.MergeSetBlues()) == 1 && ghostdagData.MergeSetBlues()[0].Equal(model.VirtualGenesisBlockHash) {
		var err error
		ghostdagData, err = c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, true)
		if err != nil {
			return nil, err
		}
	}

	mergeSet := append(ghostdagData.MergeSetBlues(), ghostdagData.MergeSetReds()...)
	mergeSetBlocks, err := c.blockStore.Blocks(c.databaseContext, stagingArea, mergeSet)
	if err != nil {
		return nil, err
	}

	mergeSetSubsidySum := int64(0)
	for _, mergeSetBlock := range mergeSetBlocks {
		coinbaseTransaction := mergeSetBlock.Transactions[transactionhelper.CoinbaseTransactionIndex]
		_, _, mergeSetBlockSubsidy, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(coinbaseTransaction)
		if err != nil {
			return nil, err
		}
		mergeSetSubsidySum += int64(mergeSetBlockSubsidy)
	}
	return big.NewRat(mergeSetSubsidySum, 1), nil
}

func (c *coinbaseManager) calculateSubsidyRandomVariable(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (int64, error) {
	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return 0, err
	}
	selectedParent := ghostdagData.SelectedParent()
	if selectedParent == nil {
		return 0, nil
	}

	seed := int64(0)
	for i := 0; i < externalapi.DomainHashSize; i += 8 {
		seed += int64(binary.LittleEndian.Uint64(selectedParent.ByteSlice()[i : i+8]))
	}
	random := rand.New(rand.NewSource(seed))

	const binomialSteps = 10
	binomialSum := int64(0)
	for i := 0; i < binomialSteps; i++ {
		step := random.Intn(2)
		binomialSum += int64(step)
	}
	return binomialSum - (binomialSteps / 2), nil
}

// Adapted from https://stackoverflow.com/a/101613
func powInt64(base int64, exponent int64) int64 {
	if exponent < 0 {
		panic("negative exponents are not supported")
	}

	result := int64(1)
	for exponent != 0 {
		if exponent&1 == 1 {
			result *= base
		}
		exponent >>= 1
		base *= base
	}
	return result
}

func (c *coinbaseManager) calcMergedBlockReward(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	blockAcceptanceData *externalapi.BlockAcceptanceData, mergingBlockDAAAddedBlocksSet hashset.HashSet) (uint64, uint64, error) {

	if !blockHash.Equal(blockAcceptanceData.BlockHash) {
		return 0, 0, errors.Errorf("blockAcceptanceData.BlockHash is expected to be %s but got %s",
			blockHash, blockAcceptanceData.BlockHash)
	}

	if !mergingBlockDAAAddedBlocksSet.Contains(blockHash) {
		return 0, 0, nil
	}

	totalFees := uint64(0)
	for _, txAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	subsidy, err := c.calcBlockSubsidy(stagingArea, blockHash)
	if err != nil {
		return 0, 0, err
	}

	return subsidy, totalFees, nil
}

// New instantiates a new CoinbaseManager
func New(
	databaseContext model.DBReader,

	subsidyGenesisReward uint64,
	minSubsidy uint64,
	maxSubsidy uint64,
	subsidyPastRewardMultiplier *big.Rat,
	subsidyMergeSetRewardMultiplier *big.Rat,
	coinbasePayloadScriptPublicKeyMaxLength uint8,
	genesisHash *externalapi.DomainHash,

	dagTraversalManager model.DAGTraversalManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore,
	daaBlocksStore model.DAABlocksStore,
	blockStore model.BlockStore) model.CoinbaseManager {

	return &coinbaseManager{
		databaseContext: databaseContext,

		subsidyGenesisReward:                    subsidyGenesisReward,
		minSubsidy:                              minSubsidy,
		maxSubsidy:                              maxSubsidy,
		subsidyPastRewardMultiplier:             subsidyPastRewardMultiplier,
		subsidyMergeSetRewardMultiplier:         subsidyMergeSetRewardMultiplier,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,
		genesisHash:                             genesisHash,

		dagTraversalManager: dagTraversalManager,
		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
		daaBlocksStore:      daaBlocksStore,
		blockStore:          blockStore,
	}
}
