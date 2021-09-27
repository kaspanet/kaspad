package coinbasemanager

import (
	"encoding/binary"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"math/rand"
)

type coinbaseManager struct {
	subsidyGenesisReward                    uint64
	subsidyPastRewardMultiplier             *big.Rat
	subsidyMergeSetRewardMultiplier         *big.Rat
	coinbasePayloadScriptPublicKeyMaxLength uint8
	genesisHash                             *externalapi.DomainHash

	databaseContext     model.DBReader
	dagTraversalManager model.DAGTraversalManager
	dagTopologyManager  model.DAGTopologyManager
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
	daaBlocksStore      model.DAABlocksStore
	subsidyStore        model.SubsidyStore
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
	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	for _, blue := range ghostdagData.MergeSetBlues() {
		txOut, hasReward, err := c.coinbaseOutputForBlueBlock(stagingArea, blue, acceptanceDataMap[*blue], daaAddedBlocksSet)
		if err != nil {
			return nil, err
		}

		if hasReward {
			txOuts = append(txOuts, txOut)
		}
	}

	txOut, hasReward, err := c.coinbaseOutputForRewardFromRedBlocks(
		stagingArea, ghostdagData, acceptanceData, daaAddedBlocksSet, coinbaseData)
	if err != nil {
		return nil, err
	}

	if hasReward {
		txOuts = append(txOuts, txOut)
	}

	payload, err := c.serializeCoinbasePayload(ghostdagData.BlueScore(), coinbaseData)
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

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func (c *coinbaseManager) coinbaseOutputForBlueBlock(stagingArea *model.StagingArea,
	blueBlock *externalapi.DomainHash, blockAcceptanceData *externalapi.BlockAcceptanceData,
	mergingBlockDAAAddedBlocksSet hashset.HashSet) (*externalapi.DomainTransactionOutput, bool, error) {

	totalReward, err := c.calcMergedBlockReward(stagingArea, blueBlock, blockAcceptanceData, mergingBlockDAAAddedBlocksSet)
	if err != nil {
		return nil, false, err
	}

	if totalReward == 0 {
		return nil, false, nil
	}

	// the ScriptPublicKey for the coinbase is parsed from the coinbase payload
	_, coinbaseData, err := c.ExtractCoinbaseDataAndBlueScore(blockAcceptanceData.TransactionAcceptanceData[0].Transaction)
	if err != nil {
		return nil, false, err
	}

	txOut := &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}

	return txOut, true, nil
}

func (c *coinbaseManager) coinbaseOutputForRewardFromRedBlocks(stagingArea *model.StagingArea,
	ghostdagData *externalapi.BlockGHOSTDAGData, acceptanceData externalapi.AcceptanceData, daaAddedBlocksSet hashset.HashSet,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransactionOutput, bool, error) {

	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	totalReward := uint64(0)
	for _, red := range ghostdagData.MergeSetReds() {
		reward, err := c.calcMergedBlockReward(stagingArea, red, acceptanceDataMap[*red], daaAddedBlocksSet)
		if err != nil {
			return nil, false, err
		}

		totalReward += reward
	}

	if totalReward == 0 {
		return nil, false, nil
	}

	return &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}, true, nil
}

func acceptanceDataFromArrayToMap(acceptanceData externalapi.AcceptanceData) map[externalapi.DomainHash]*externalapi.BlockAcceptanceData {
	acceptanceDataMap := make(map[externalapi.DomainHash]*externalapi.BlockAcceptanceData, len(acceptanceData))
	for _, blockAcceptanceData := range acceptanceData {
		acceptanceDataMap[*blockAcceptanceData.BlockHash] = blockAcceptanceData
	}
	return acceptanceDataMap
}

func (c *coinbaseManager) getBlockSubsidy(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	subsidyExists, err := c.subsidyStore.Has(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return 0, err
	}
	if subsidyExists {
		return c.subsidyStore.Get(c.databaseContext, stagingArea, blockHash)
	}

	subsidy, err := c.calcBlockSubsidy(stagingArea, blockHash)
	if err != nil {
		return 0, err
	}
	c.subsidyStore.Stage(stagingArea, blockHash, subsidy)
	return subsidy, nil
}

// calcBlockSubsidy returns the subsidy amount a block at the provided blue score
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// Further details: https://hashdag.medium.com/kaspa-launch-plan-9a63f4d754a6
//
// TODO: This function makes heavy use of floating point operations, which are
// unfortunately not guaranteed to produce identical results between differing
// architectures. This may produce discrepancies among nodes in the network
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

	fmt.Println("----------")
	fmt.Println(blockHash.String(), averagePastSubsidy, subsidyRandomVariable, mergeSetSubsidySum)

	pastSubsidy := new(big.Rat).Mul(averagePastSubsidy, c.subsidyPastRewardMultiplier)
	subsidyRandom := new(big.Rat).SetFloat64(math.Pow(4, subsidyRandomVariable))
	mergeSetSubsidy := new(big.Rat).Mul(mergeSetSubsidySum, c.subsidyMergeSetRewardMultiplier)

	fmt.Println(pastSubsidy, subsidyRandom, mergeSetSubsidy)

	blockSubsidyBigRat := new(big.Rat).Add(mergeSetSubsidy, new(big.Rat).Mul(pastSubsidy, subsidyRandom))
	blockSubsidyFloat64, _ := blockSubsidyBigRat.Float64()
	blockSubsidyUint64 := uint64(blockSubsidyFloat64)

	fmt.Println(blockSubsidyBigRat, blockSubsidyFloat64, blockSubsidyUint64)
	fmt.Println("============")

	return blockSubsidyUint64, nil
}

func (c *coinbaseManager) calculateAveragePastSubsidy(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*big.Rat, error) {
	blockParents, err := c.dagTopologyManager.Parents(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	if len(blockParents) == 0 {
		return nil, nil
	}

	pastBlockCount := int64(0)
	pastBlockSubsidySum := int64(0)
	queue := c.dagTraversalManager.NewDownHeap(stagingArea)
	addedToQueue := make(map[externalapi.DomainHash]struct{})

	err = queue.PushSlice(blockParents)
	if err != nil {
		return nil, err
	}
	for _, blockParent := range blockParents {
		addedToQueue[*blockParent] = struct{}{}
	}

	const subsidyPastWindowSize = int64(100)
	for pastBlockCount < subsidyPastWindowSize && queue.Len() > 0 {
		pastBlockHash := queue.Pop()
		pastBlockCount++

		pastBlockSubsidy, err := c.getBlockSubsidy(stagingArea, pastBlockHash)
		if err != nil {
			return nil, err
		}
		pastBlockSubsidySum += int64(pastBlockSubsidy)

		pastBlockParents, err := c.dagTopologyManager.Parents(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}
		for _, pastBlockParent := range pastBlockParents {
			if _, ok := addedToQueue[*pastBlockParent]; ok {
				continue
			}
			err = queue.Push(pastBlockParent)
			if err != nil {
				return nil, err
			}
			addedToQueue[*pastBlockParent] = struct{}{}
		}
	}

	return big.NewRat(pastBlockSubsidySum, pastBlockCount), nil
}

func (c *coinbaseManager) calculateMergeSetSubsidySum(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*big.Rat, error) {
	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}
	mergeSet := append(ghostdagData.MergeSetBlues(), ghostdagData.MergeSetReds()...)

	mergeSetSubsidySum := int64(0)
	for _, mergeSetBlockHash := range mergeSet {
		mergeSetBlockSubsidy, err := c.getBlockSubsidy(stagingArea, mergeSetBlockHash)
		if err != nil {
			return nil, err
		}
		mergeSetSubsidySum += int64(mergeSetBlockSubsidy)
	}
	return big.NewRat(mergeSetSubsidySum, 1), nil
}

func (c *coinbaseManager) calculateSubsidyRandomVariable(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (float64, error) {
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
	randomNormalFloat64 := random.NormFloat64()
	return randomNormalFloat64, nil
}

func (c *coinbaseManager) calcMergedBlockReward(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	blockAcceptanceData *externalapi.BlockAcceptanceData, mergingBlockDAAAddedBlocksSet hashset.HashSet) (uint64, error) {

	if !blockHash.Equal(blockAcceptanceData.BlockHash) {
		return 0, errors.Errorf("blockAcceptanceData.BlockHash is expected to be %s but got %s",
			blockHash, blockAcceptanceData.BlockHash)
	}

	if !mergingBlockDAAAddedBlocksSet.Contains(blockHash) {
		return 0, nil
	}

	totalFees := uint64(0)
	for _, txAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	subsidy, err := c.getBlockSubsidy(stagingArea, blockHash)
	if err != nil {
		return 0, err
	}

	return subsidy + totalFees, nil
}

// New instantiates a new CoinbaseManager
func New(
	databaseContext model.DBReader,

	subsidyGenesisReward uint64,
	subsidyPastRewardMultiplier *big.Rat,
	subsidyMergeSetRewardMultiplier *big.Rat,
	coinbasePayloadScriptPublicKeyMaxLength uint8,
	genesisHash *externalapi.DomainHash,

	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore,
	daaBlocksStore model.DAABlocksStore,
	subsidyStore model.SubsidyStore) model.CoinbaseManager {

	return &coinbaseManager{
		databaseContext: databaseContext,

		subsidyGenesisReward:                    subsidyGenesisReward,
		subsidyPastRewardMultiplier:             subsidyPastRewardMultiplier,
		subsidyMergeSetRewardMultiplier:         subsidyMergeSetRewardMultiplier,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,
		genesisHash:                             genesisHash,

		dagTraversalManager: dagTraversalManager,
		dagTopologyManager:  dagTopologyManager,
		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
		daaBlocksStore:      daaBlocksStore,
		subsidyStore:        subsidyStore,
	}
}
