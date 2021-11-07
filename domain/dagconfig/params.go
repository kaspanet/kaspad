// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math/big"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/network"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util"
)

// These variables are the DAG proof-of-work limit parameters for each default
// network.
var (
	// bigOne is 1 represented as a big.Int. It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// mainPowMax is the highest proof of work value a Kaspa block can
	// have for the main network. It is the value 2^255 - 1.
	mainPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// testnetPowMax is the highest proof of work value a Kaspa block
	// can have for the test network. It is the value 2^255 - 1.
	testnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// simnetPowMax is the highest proof of work value a Kaspa block
	// can have for the simulation test network. It is the value 2^255 - 1.
	simnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// devnetPowMax is the highest proof of work value a Kaspa block
	// can have for the development network. It is the value
	// 2^255 - 1.
	devnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)
)

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType uint8

// Params defines a Kaspa network by its parameters. These parameters may be
// used by Kaspa applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	// K defines the K parameter for GHOSTDAG consensus algorithm.
	// See ghostdag.go for further details.
	K externalapi.KType

	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net appmessage.KaspaNet

	// RPCPort defines the rpc server port
	RPCPort string

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string

	// DNSSeeds defines a list of DNS seeds for the network that are used
	// as one method to discover peers.
	DNSSeeds []string

	// GRPCSeeds defines a list of GRPC seeds for the network that are used
	// as one method to discover peers.
	GRPCSeeds []string

	// GenesisBlock defines the first block of the DAG.
	GenesisBlock *externalapi.DomainBlock

	// GenesisHash is the starting block hash.
	GenesisHash *externalapi.DomainHash

	// PowMax defines the highest allowed proof of work value for a block
	// as a uint256.
	PowMax *big.Int

	// BlockCoinbaseMaturity is the number of blocks required before newly mined
	// coins can be spent.
	BlockCoinbaseMaturity uint64

	// SubsidyGenesisReward SubsidyMergeSetRewardMultiplier, and
	// SubsidyPastRewardMultiplier are part of the block subsidy equation.
	// Further details: https://hashdag.medium.com/kaspa-launch-plan-9a63f4d754a6
	SubsidyGenesisReward            uint64
	MinSubsidy                      uint64
	MaxSubsidy                      uint64
	SubsidyPastRewardMultiplier     *big.Rat
	SubsidyMergeSetRewardMultiplier *big.Rat

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// FinalityDuration is the duration of the finality window.
	FinalityDuration time.Duration

	// TimestampDeviationTolerance is the maximum offset a block timestamp
	// is allowed to be in the future before it gets delayed
	TimestampDeviationTolerance int

	// DifficultyAdjustmentWindowSize is the size of window that is inspected
	// to calculate the required difficulty of each block.
	DifficultyAdjustmentWindowSize int

	// These fields are related to voting on consensus rule changes as
	// defined by BIP0009.
	//
	// RuleChangeActivationThreshold is the number of blocks in a threshold
	// state retarget window for which a positive vote for a rule change
	// must be cast in order to lock in a rule change. It should typically
	// be 95% for the main network and 75% for test networks.
	//
	// MinerConfirmationWindow is the number of blocks in each threshold
	// state retarget window.
	//
	// Deployments define the specific consensus rule changes to be voted
	// on.
	RuleChangeActivationThreshold uint64
	MinerConfirmationWindow       uint64

	// Mempool parameters
	RelayNonStdTxs bool

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable bool

	// Human-readable prefix for Bech32 encoded addresses
	Prefix util.Bech32Prefix

	// Address encoding magics
	PrivateKeyID byte // First byte of a WIF private key

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks bool

	// DisableDifficultyAdjustment determine whether to use difficulty
	DisableDifficultyAdjustment bool

	// SkipProofOfWork indicates whether proof of work should be checked.
	SkipProofOfWork bool

	// MaxCoinbasePayloadLength is the maximum length in bytes allowed for a block's coinbase's payload
	MaxCoinbasePayloadLength uint64

	// MaxBlockMass is the maximum mass a block is allowed
	MaxBlockMass uint64

	// MaxBlockParents is the maximum number of blocks a block is allowed to point to
	MaxBlockParents externalapi.KType

	// MassPerTxByte is the number of grams that any byte
	// adds to a transaction.
	MassPerTxByte uint64

	// MassPerScriptPubKeyByte is the number of grams that any
	// scriptPubKey byte adds to a transaction.
	MassPerScriptPubKeyByte uint64

	// MassPerSigOp is the number of grams that any
	// signature operation adds to a transaction.
	MassPerSigOp uint64

	// MergeSetSizeLimit is the maximum number of blocks in a block's merge set
	MergeSetSizeLimit uint64

	// CoinbasePayloadScriptPublicKeyMaxLength is the maximum allowed script public key in the coinbase's payload
	CoinbasePayloadScriptPublicKeyMaxLength uint8

	// PruningProofM is the 'm' constant in the pruning proof. For more details see: https://github.com/kaspanet/research/issues/3
	PruningProofM uint64

	// BaseSubsidy is the starting subsidy amount for mined blocks.
	BaseSubsidy uint64

	FixedSubsidySwitchPruningPointInterval uint64

	FixedSubsidySwitchHashRateThreshold *big.Int
}

// NormalizeRPCServerAddress returns addr with the current network default
// port appended if there is not already a port specified.
func (p *Params) NormalizeRPCServerAddress(addr string) (string, error) {
	return network.NormalizeAddress(addr, p.RPCPort)
}

// FinalityDepth returns the finality duration represented in blocks
func (p *Params) FinalityDepth() uint64 {
	return uint64(p.FinalityDuration / p.TargetTimePerBlock)
}

// PruningDepth returns the pruning duration represented in blocks
func (p *Params) PruningDepth() uint64 {
	return 2*p.FinalityDepth() + 4*p.MergeSetSizeLimit*uint64(p.K) + 2*uint64(p.K) + 2
}

// MainnetParams defines the network parameters for the main Kaspa network.
var MainnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "kaspa-mainnet",
	Net:         appmessage.Mainnet,
	RPCPort:     "16110",
	DefaultPort: "16111",
	DNSSeeds:    []string{"mainnet-dnsseed.daglabs-dev.com"},

	// DAG parameters
	GenesisBlock:                    &genesisBlock,
	GenesisHash:                     genesisHash,
	PowMax:                          mainPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	MinSubsidy:                      defaultMinSubsidy,
	MaxSubsidy:                      defaultMaxSubsidy,
	SubsidyPastRewardMultiplier:     big.NewRat(9, 10),
	SubsidyMergeSetRewardMultiplier: big.NewRat(1, 10),
	TargetTimePerBlock:              defaultTargetTimePerBlock,
	FinalityDuration:                defaultFinalityDuration,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1916, // 95% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016, //

	// Mempool parameters
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixKaspa,

	// Address encoding magics
	PrivateKeyID: 0x80, // starts with 5 (uncompressed) or K (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: false,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	FixedSubsidySwitchPruningPointInterval:  defaultFixedSubsidySwitchPruningPointInterval,
	FixedSubsidySwitchHashRateThreshold:     big.NewInt(150_000_000_000),
}

// TestnetParams defines the network parameters for the test Kaspa network.
var TestnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "kaspa-testnet-6",
	Net:         appmessage.Testnet,
	RPCPort:     "16210",
	DefaultPort: "16211",
	DNSSeeds:    []string{"testnet-6-dnsseed.daglabs-dev.com"},

	// DAG parameters
	GenesisBlock:                    &testnetGenesisBlock,
	GenesisHash:                     testnetGenesisHash,
	PowMax:                          testnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	MinSubsidy:                      defaultMinSubsidy,
	MaxSubsidy:                      defaultMaxSubsidy,
	SubsidyPastRewardMultiplier:     big.NewRat(9, 10),
	SubsidyMergeSetRewardMultiplier: big.NewRat(1, 10),
	TargetTimePerBlock:              defaultTargetTimePerBlock,
	FinalityDuration:                defaultFinalityDuration,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,

	// Mempool parameters
	RelayNonStdTxs: true,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixKaspaTest,

	// Address encoding magics
	PrivateKeyID: 0xef, // starts with 9 (uncompressed) or c (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: false,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	FixedSubsidySwitchPruningPointInterval:  defaultFixedSubsidySwitchPruningPointInterval,
	FixedSubsidySwitchHashRateThreshold:     big.NewInt(150_000_000_000),
}

// SimnetParams defines the network parameters for the simulation test Kaspa
// network. This network is similar to the normal test network except it is
// intended for private use within a group of individuals doing simulation
// testing. The functionality is intended to differ in that the only nodes
// which are specifically specified are used to create the network rather than
// following normal discovery rules. This is important as otherwise it would
// just turn into another public testnet.
var SimnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "kaspa-simnet",
	Net:         appmessage.Simnet,
	RPCPort:     "16510",
	DefaultPort: "16511",
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                    &simnetGenesisBlock,
	GenesisHash:                     simnetGenesisHash,
	PowMax:                          simnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	MinSubsidy:                      defaultMinSubsidy,
	MaxSubsidy:                      defaultMaxSubsidy,
	SubsidyPastRewardMultiplier:     big.NewRat(9, 10),
	SubsidyMergeSetRewardMultiplier: big.NewRat(1, 10),
	TargetTimePerBlock:              time.Millisecond,
	FinalityDuration:                time.Minute,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 75, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       100,

	// Mempool parameters
	RelayNonStdTxs: true,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	PrivateKeyID: 0x64, // starts with 4 (uncompressed) or F (compressed)
	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixKaspaSim,

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: true,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	FixedSubsidySwitchPruningPointInterval:  defaultFixedSubsidySwitchPruningPointInterval,
	FixedSubsidySwitchHashRateThreshold:     big.NewInt(150_000_000_000),
}

// DevnetParams defines the network parameters for the development Kaspa network.
var DevnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "kaspa-devnet",
	Net:         appmessage.Devnet,
	RPCPort:     "16610",
	DefaultPort: "16611",
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                    &devnetGenesisBlock,
	GenesisHash:                     devnetGenesisHash,
	PowMax:                          devnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	MinSubsidy:                      defaultMinSubsidy,
	MaxSubsidy:                      defaultMaxSubsidy,
	SubsidyPastRewardMultiplier:     big.NewRat(9, 10),
	SubsidyMergeSetRewardMultiplier: big.NewRat(1, 10),
	TargetTimePerBlock:              defaultTargetTimePerBlock,
	FinalityDuration:                defaultFinalityDuration,
	DifficultyAdjustmentWindowSize:  defaultDifficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:     defaultTimestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,

	// Mempool parameters
	RelayNonStdTxs: true,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: true,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixKaspaDev,

	// Address encoding magics
	PrivateKeyID: 0xef, // starts with 9 (uncompressed) or c (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,

	DisableDifficultyAdjustment: false,

	MaxCoinbasePayloadLength:                defaultMaxCoinbasePayloadLength,
	MaxBlockMass:                            defaultMaxBlockMass,
	MaxBlockParents:                         defaultMaxBlockParents,
	MassPerTxByte:                           defaultMassPerTxByte,
	MassPerScriptPubKeyByte:                 defaultMassPerScriptPubKeyByte,
	MassPerSigOp:                            defaultMassPerSigOp,
	MergeSetSizeLimit:                       defaultMergeSetSizeLimit,
	CoinbasePayloadScriptPublicKeyMaxLength: defaultCoinbasePayloadScriptPublicKeyMaxLength,
	PruningProofM:                           defaultPruningProofM,
	FixedSubsidySwitchPruningPointInterval:  defaultFixedSubsidySwitchPruningPointInterval,
	FixedSubsidySwitchHashRateThreshold:     big.NewInt(150_000_000_000),
}

var (
	// ErrDuplicateNet describes an error where the parameters for a Kaspa
	// network could not be set due to the network already being a standard
	// network or previously-registered into this package.
	ErrDuplicateNet = errors.New("duplicate Kaspa network")
)

var (
	registeredNets = make(map[appmessage.KaspaNet]struct{})
)

// Register registers the network parameters for a Kaspa network. This may
// error with ErrDuplicateNet if the network is already registered (either
// due to a previous Register call, or the network being one of the default
// networks).
//
// Network parameters should be registered into this package by a main package
// as early as possible. Then, library packages may lookup networks or network
// parameters based on inputs and work regardless of the network being standard
// or not.
func Register(params *Params) error {
	if _, ok := registeredNets[params.Net]; ok {
		return ErrDuplicateNet
	}
	registeredNets[params.Net] = struct{}{}

	return nil
}

// mustRegister performs the same function as Register except it panics if there
// is an error. This should only be called from package init functions.
func mustRegister(params *Params) {
	if err := Register(params); err != nil {
		panic("failed to register network: " + err.Error())
	}
}

func init() {
	// Register all default networks when the package is initialized.
	mustRegister(&MainnetParams)
	mustRegister(&TestnetParams)
	mustRegister(&SimnetParams)
	mustRegister(&DevnetParams)
}
