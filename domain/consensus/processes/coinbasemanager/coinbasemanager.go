package coinbasemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
)

type coinbaseManager struct {
	subsidyReductionInterval                uint64
	baseSubsidy                             uint64
	coinbasePayloadScriptPublicKeyMaxLength uint64

	databaseContext     model.DBReader
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
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

	txOuts := make([]*externalapi.DomainTransactionOutput, 0, len(ghostdagData.MergeSetBlues()))
	for i, blue := range ghostdagData.MergeSetBlues() {
		txOut, hasReward, err := c.coinbaseOutputForBlueBlock(blue, acceptanceData[i])
		if err != nil {
			return nil, err
		}

		if hasReward {
			txOuts = append(txOuts, txOut)
		}
	}

	payload, err := c.serializeCoinbasePayload(ghostdagData.BlueScore(), coinbaseData)
	if err != nil {
		return nil, err
	}

	payloadHash := hashes.PayloadHash(payload)

	return &externalapi.DomainTransaction{
		Version:      constants.TransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      txOuts,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          0,
		PayloadHash:  *payloadHash,
		Payload:      payload,
	}, nil
}

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func (c *coinbaseManager) coinbaseOutputForBlueBlock(blueBlock *externalapi.DomainHash,
	blockAcceptanceData *model.BlockAcceptanceData) (*externalapi.DomainTransactionOutput, bool, error) {

	totalFees := uint64(0)
	for _, txAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	subsidy, err := c.calcBlockSubsidy(blueBlock)
	if err != nil {
		return nil, false, err
	}

	totalReward := subsidy + totalFees

	if totalReward == 0 {
		return nil, false, nil
	}

	// the ScriptPubKey for the coinbase is parsed from the coinbase payload
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

	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, blockHash)
	if err != nil {
		return 0, err
	}

	// Equivalent to: baseSubsidy / 2^(blueScore/subsidyHalvingInterval)
	return c.baseSubsidy >> uint(ghostdagData.BlueScore()/c.subsidyReductionInterval), nil
}

// New instantiates a new CoinbaseManager
func New(
	databaseContext model.DBReader,

	subsidyReductionInterval uint64,
	baseSubsidy uint64,
	coinbasePayloadScriptPublicKeyMaxLength uint64,

	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore) model.CoinbaseManager {

	return &coinbaseManager{
		databaseContext: databaseContext,

		subsidyReductionInterval:                subsidyReductionInterval,
		baseSubsidy:                             baseSubsidy,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,

		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
	}
}
