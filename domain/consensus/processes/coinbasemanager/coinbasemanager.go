package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
)

type coinbaseManager struct {
	subsidyReductionInterval                uint64
	baseSubsidy                             uint64
	coinbasePayloadScriptPublicKeyMaxLength uint8

	databaseContext     model.DBReader
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
	daaBlocksStore      model.DAABlocksStore
}

func (c *coinbaseManager) ExpectedCoinbaseTransaction(blockHash *externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransaction, error) {

	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	acceptanceData, err := c.acceptanceDataStore.Get(c.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	daaAddedBlocksSet, err := c.daaAddedBlocksSet(blockHash)
	if err != nil {
		return nil, err
	}

	txOuts := make([]*externalapi.DomainTransactionOutput, 0, len(ghostdagData.MergeSetBlues()))
	for i, blue := range ghostdagData.MergeSetBlues() {
		txOut, hasReward, err := c.coinbaseOutputForBlueBlock(blue, acceptanceData[i], daaAddedBlocksSet)
		if err != nil {
			return nil, err
		}

		if hasReward {
			txOuts = append(txOuts, txOut)
		}
	}

	txOut, hasReward, err := c.coinbaseOutputForRewardFromRedBlocks(ghostdagData, acceptanceData, daaAddedBlocksSet, coinbaseData)
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

func (c *coinbaseManager) daaAddedBlocksSet(blockHash *externalapi.DomainHash) (hashset.HashSet, error) {
	daaAddedBlocks, err := c.daaBlocksStore.DAAAddedBlocks(c.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	daaAddedBlocksSet := hashset.New()
	for _, block := range daaAddedBlocks {
		daaAddedBlocksSet.Add(block)
	}
	return daaAddedBlocksSet, nil
}

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func (c *coinbaseManager) coinbaseOutputForBlueBlock(blueBlock *externalapi.DomainHash,
	blockAcceptanceData *externalapi.BlockAcceptanceData,
	mergingBlockDAAAddedBlocksSet hashset.HashSet) (*externalapi.DomainTransactionOutput, bool, error) {

	totalReward, err := c.calcMergedBlockReward(blueBlock, blockAcceptanceData, mergingBlockDAAAddedBlocksSet)
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

func (c *coinbaseManager) coinbaseOutputForRewardFromRedBlocks(ghostdagData *model.BlockGHOSTDAGData,
	acceptanceData externalapi.AcceptanceData, daaAddedBlocksSet hashset.HashSet,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransactionOutput, bool, error) {

	totalReward := uint64(0)
	mergeSetBluesCount := len(ghostdagData.MergeSetBlues())
	for i, red := range ghostdagData.MergeSetReds() {
		reward, err := c.calcMergedBlockReward(red, acceptanceData[mergeSetBluesCount+i], daaAddedBlocksSet)
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

// calcBlockSubsidy returns the subsidy amount a block at the provided blue score
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// The subsidy is halved every SubsidyReductionInterval blocks. Mathematically
// this is: baseSubsidy / 2^(blueScore/SubsidyReductionInterval)
//
// At the target block generation rate for the main network, this is
// approximately every 4 years.
func (c *coinbaseManager) calcBlockSubsidy(blockHash *externalapi.DomainHash) (uint64, error) {
	if c.subsidyReductionInterval == 0 {
		return c.baseSubsidy, nil
	}

	daaScore, err := c.daaBlocksStore.DAAScore(c.databaseContext, blockHash)
	if err != nil {
		return 0, err
	}

	// Equivalent to: baseSubsidy / 2^(daaScore/subsidyHalvingInterval)
	return c.baseSubsidy >> uint(daaScore/c.subsidyReductionInterval), nil
}

func (c *coinbaseManager) calcMergedBlockReward(blockHash *externalapi.DomainHash,
	blockAcceptanceData *externalapi.BlockAcceptanceData, mergingBlockDAAAddedBlocksSet hashset.HashSet) (uint64, error) {

	if !mergingBlockDAAAddedBlocksSet.Contains(blockHash) {
		return 0, nil
	}

	totalFees := uint64(0)
	for _, txAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	subsidy, err := c.calcBlockSubsidy(blockHash)
	if err != nil {
		return 0, err
	}

	return subsidy + totalFees, nil
}

// New instantiates a new CoinbaseManager
func New(
	databaseContext model.DBReader,

	subsidyReductionInterval uint64,
	baseSubsidy uint64,
	coinbasePayloadScriptPublicKeyMaxLength uint8,

	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore,
	daaBlocksStore model.DAABlocksStore) model.CoinbaseManager {

	return &coinbaseManager{
		databaseContext: databaseContext,

		subsidyReductionInterval:                subsidyReductionInterval,
		baseSubsidy:                             baseSubsidy,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,

		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
		daaBlocksStore:      daaBlocksStore,
	}
}
