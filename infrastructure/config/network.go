package config

import (
	"encoding/json"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/pkg/errors"
)

// NetworkFlags holds the network configuration, that is which network is selected.
type NetworkFlags struct {
	Testnet               bool   `long:"testnet" description:"Use the test network"`
	Simnet                bool   `long:"simnet" description:"Use the simulation test network"`
	Devnet                bool   `long:"devnet" description:"Use the development test network"`
	OverrideDAGParamsFile string `long:"override-dag-params-file" description:"Overrides DAG params (allowed only on devnet)"`

	ActiveNetParams *dagconfig.Params
}

type overrideDAGParamsConfig struct {
	K                                       *externalapi.KType `json:"k"`
	MaxBlockParents                         *externalapi.KType `json:"maxBlockParents"`
	MergeSetSizeLimit                       *uint64            `json:"mergeSetSizeLimit"`
	MaxBlockMass                            *uint64            `json:"maxBlockMass"`
	MaxCoinbasePayloadLength                *uint64            `json:"maxCoinbasePayloadLength"`
	MassPerTxByte                           *uint64            `json:"massPerTxByte"`
	MassPerScriptPubKeyByte                 *uint64            `json:"massPerScriptPubKeyByte"`
	MassPerSigOp                            *uint64            `json:"massPerSigOp"`
	CoinbasePayloadScriptPublicKeyMaxLength *uint8             `json:"coinbasePayloadScriptPublicKeyMaxLength"`
	PowMax                                  *string            `json:"powMax"`
	BlockCoinbaseMaturity                   *uint64            `json:"blockCoinbaseMaturity"`
	SubsidyGenesisReward                    *uint64            `json:"subsidyGenesisReward"`
	SubsidyPastRewardMultiplier             *float64           `json:"subsidyPastRewardMultiplier"`
	SubsidyMergeSetRewardMultiplier         *float64           `json:"subsidyMergeSetRewardMultiplier"`
	TargetTimePerBlockInMilliSeconds        *int64             `json:"targetTimePerBlockInMilliSeconds"`
	FinalityDuration                        *int64             `json:"finalityDuration"`
	TimestampDeviationTolerance             *int               `json:"timestampDeviationTolerance"`
	DifficultyAdjustmentWindowSize          *int               `json:"difficultyAdjustmentWindowSize"`
	RelayNonStdTxs                          *bool              `json:"relayNonStdTxs"`
	AcceptUnroutable                        *bool              `json:"acceptUnroutable"`
	EnableNonNativeSubnetworks              *bool              `json:"enableNonNativeSubnetworks"`
	DisableDifficultyAdjustment             *bool              `json:"disableDifficultyAdjustment"`
	SkipProofOfWork                         *bool              `json:"skipProofOfWork"`
	HardForkOmitGenesisFromParentsDAAScore  *uint64            `json:"hardForkOmitGenesisFromParentsDaaScore"`
}

// ResolveNetwork parses the network command line argument and sets NetParams accordingly.
// It returns error if more than one network was selected, nil otherwise.
func (networkFlags *NetworkFlags) ResolveNetwork(parser *flags.Parser) error {
	//NetParams holds the selected network parameters. Default value is main-net.
	networkFlags.ActiveNetParams = &dagconfig.MainnetParams
	// Multiple networks can't be selected simultaneously.
	numNets := 0
	// default net is main net
	// Count number of network flags passed; assign active network params
	// while we're at it
	if networkFlags.Testnet {
		numNets++
		networkFlags.ActiveNetParams = &dagconfig.TestnetParams
	}
	if networkFlags.Simnet {
		numNets++
		networkFlags.ActiveNetParams = &dagconfig.SimnetParams
	}
	if networkFlags.Devnet {
		numNets++
		networkFlags.ActiveNetParams = &dagconfig.DevnetParams
	}
	if numNets > 1 {
		message := "Multiple networks parameters (testnet, simnet, devnet, etc.) cannot be used" +
			"together. Please choose only one network"
		err := errors.Errorf(message)
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return err
	}

	err := networkFlags.overrideDAGParams()
	if err != nil {
		return err
	}

	return nil
}

// NetParams returns the ActiveNetParams
func (networkFlags *NetworkFlags) NetParams() *dagconfig.Params {
	return networkFlags.ActiveNetParams
}

func (networkFlags *NetworkFlags) overrideDAGParams() error {

	if networkFlags.OverrideDAGParamsFile == "" {
		return nil
	}

	if !networkFlags.Devnet {
		return errors.Errorf("override-dag-params-file is allowed only when using devnet")
	}

	overrideDAGParamsFile, err := os.Open(networkFlags.OverrideDAGParamsFile)
	if err != nil {
		return err
	}
	defer overrideDAGParamsFile.Close()

	decoder := json.NewDecoder(overrideDAGParamsFile)
	config := &overrideDAGParamsConfig{}
	err = decoder.Decode(config)
	if err != nil {
		return err
	}

	if config.K != nil {
		networkFlags.ActiveNetParams.K = *config.K
	}

	if config.MaxBlockParents != nil {
		networkFlags.ActiveNetParams.MaxBlockParents = *config.MaxBlockParents
	}

	if config.MergeSetSizeLimit != nil {
		networkFlags.ActiveNetParams.MergeSetSizeLimit = *config.MergeSetSizeLimit
	}

	if config.MaxBlockMass != nil {
		networkFlags.ActiveNetParams.MaxBlockMass = *config.MaxBlockMass
	}

	if config.MaxCoinbasePayloadLength != nil {
		networkFlags.ActiveNetParams.MaxCoinbasePayloadLength = *config.MaxCoinbasePayloadLength
	}

	if config.MassPerTxByte != nil {
		networkFlags.ActiveNetParams.MassPerTxByte = *config.MassPerTxByte
	}

	if config.MassPerScriptPubKeyByte != nil {
		networkFlags.ActiveNetParams.MassPerScriptPubKeyByte = *config.MassPerScriptPubKeyByte
	}

	if config.MassPerSigOp != nil {
		networkFlags.ActiveNetParams.MassPerSigOp = *config.MassPerSigOp
	}

	if config.CoinbasePayloadScriptPublicKeyMaxLength != nil {
		networkFlags.ActiveNetParams.CoinbasePayloadScriptPublicKeyMaxLength = *config.CoinbasePayloadScriptPublicKeyMaxLength
	}

	if config.PowMax != nil {
		powMax, ok := big.NewInt(0).SetString(*config.PowMax, 16)
		if !ok {
			return errors.Errorf("couldn't convert %s to big int", *config.PowMax)
		}

		genesisTarget := difficulty.CompactToBig(networkFlags.ActiveNetParams.GenesisBlock.Header.Bits())
		if powMax.Cmp(genesisTarget) > 0 {
			return errors.Errorf("powMax (%s) is smaller than genesis's target (%s)", powMax.Text(16),
				genesisTarget.Text(16))
		}
		networkFlags.ActiveNetParams.PowMax = powMax
	}

	if config.BlockCoinbaseMaturity != nil {
		networkFlags.ActiveNetParams.BlockCoinbaseMaturity = *config.BlockCoinbaseMaturity
	}

	if config.SubsidyGenesisReward != nil {
		networkFlags.ActiveNetParams.SubsidyGenesisReward = *config.SubsidyGenesisReward
	}

	if config.TargetTimePerBlockInMilliSeconds != nil {
		networkFlags.ActiveNetParams.TargetTimePerBlock = time.Duration(*config.TargetTimePerBlockInMilliSeconds) *
			time.Millisecond
	}

	if config.FinalityDuration != nil {
		networkFlags.ActiveNetParams.FinalityDuration = time.Duration(*config.FinalityDuration) * time.Millisecond
	}

	if config.TimestampDeviationTolerance != nil {
		networkFlags.ActiveNetParams.TimestampDeviationTolerance = *config.TimestampDeviationTolerance
	}

	if config.DifficultyAdjustmentWindowSize != nil {
		networkFlags.ActiveNetParams.DifficultyAdjustmentWindowSize = *config.DifficultyAdjustmentWindowSize
	}

	if config.TimestampDeviationTolerance != nil {
		networkFlags.ActiveNetParams.TimestampDeviationTolerance = *config.TimestampDeviationTolerance
	}

	if config.RelayNonStdTxs != nil {
		networkFlags.ActiveNetParams.RelayNonStdTxs = *config.RelayNonStdTxs
	}

	if config.AcceptUnroutable != nil {
		networkFlags.ActiveNetParams.AcceptUnroutable = *config.AcceptUnroutable
	}

	if config.EnableNonNativeSubnetworks != nil {
		networkFlags.ActiveNetParams.EnableNonNativeSubnetworks = *config.EnableNonNativeSubnetworks
	}

	if config.SkipProofOfWork != nil {
		networkFlags.ActiveNetParams.SkipProofOfWork = *config.SkipProofOfWork
	}

	return nil
}
