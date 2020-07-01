// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"github.com/kaspanet/kaspad/util/network"
	"math"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
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

	// regressionPowMax is the highest proof of work value a Kaspa block
	// can have for the regression test network. It is the value 2^255 - 1.
	regressionPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// testnetPowMax is the highest proof of work value a Kaspa block
	// can have for the test network. It is the value 2^239 - 1.
	testnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 239), bigOne)

	// simnetPowMax is the highest proof of work value a Kaspa block
	// can have for the simulation test network. It is the value 2^255 - 1.
	simnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// devnetPowMax is the highest proof of work value a Kaspa block
	// can have for the development network. It is the value
	// 2^239 - 1.
	devnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 239), bigOne)
)

const (
	ghostdagK                      = 15
	difficultyAdjustmentWindowSize = 2640
	timestampDeviationTolerance    = 132
	finalityDuration               = 24 * time.Hour
	targetTimePerBlock             = 1 * time.Second
	finalityInterval               = uint64(finalityDuration / targetTimePerBlock)
)

// ConsensusDeployment defines details related to a specific consensus rule
// change that is voted in. This is part of BIP0009.
type ConsensusDeployment struct {
	// BitNumber defines the specific bit number within the block version
	// this particular soft-fork deployment refers to.
	BitNumber uint8

	// StartTime is the median block time after which voting on the
	// deployment starts.
	StartTime uint64

	// ExpireTime is the median block time after which the attempted
	// deployment expires.
	ExpireTime uint64
}

// Constants that define the deployment offset in the deployments field of the
// parameters for each deployment. This is useful to be able to get the details
// of a specific deployment by name.
const (
	// DeploymentTestDummy defines the rule change deployment ID for testing
	// purposes.
	DeploymentTestDummy = iota

	// NOTE: DefinedDeployments must always come last since it is used to
	// determine how many defined deployments there currently are.

	// DefinedDeployments is the number of currently defined deployments.
	DefinedDeployments
)

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType uint8

// Params defines a Kaspa network by its parameters. These parameters may be
// used by Kaspa applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	// K defines the K parameter for GHOSTDAG consensus algorithm.
	// See ghostdag.go for further details.
	K KType

	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net wire.KaspaNet

	// RPCPort defines the rpc server port
	RPCPort string

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string

	// DNSSeeds defines a list of DNS seeds for the network that are used
	// as one method to discover peers.
	DNSSeeds []string

	// GenesisBlock defines the first block of the DAG.
	GenesisBlock *wire.MsgBlock

	// GenesisHash is the starting block hash.
	GenesisHash *daghash.Hash

	// PowMax defines the highest allowed proof of work value for a block
	// as a uint256.
	PowMax *big.Int

	// BlockCoinbaseMaturity is the number of blocks required before newly mined
	// coins can be spent.
	BlockCoinbaseMaturity uint64

	// SubsidyReductionInterval is the interval of blocks before the subsidy
	// is reduced.
	SubsidyReductionInterval uint64

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// FinalityInterval is the interval that determines the finality window of the DAG.
	FinalityInterval uint64

	// TimestampDeviationTolerance is the maximum offset a block timestamp
	// is allowed to be in the future before it gets delayed
	TimestampDeviationTolerance uint64

	// DifficultyAdjustmentWindowSize is the size of window that is inspected
	// to calculate the required difficulty of each block.
	DifficultyAdjustmentWindowSize uint64

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
	Deployments                   [DefinedDeployments]ConsensusDeployment

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
}

// NormalizeRPCServerAddress returns addr with the current network default
// port appended if there is not already a port specified.
func (p *Params) NormalizeRPCServerAddress(addr string) (string, error) {
	return network.NormalizeAddress(addr, p.RPCPort)
}

// MainnetParams defines the network parameters for the main Kaspa network.
var MainnetParams = Params{
	K:           ghostdagK,
	Name:        "mainnet",
	Net:         wire.Mainnet,
	RPCPort:     "16110",
	DefaultPort: "16111",
	DNSSeeds:    []string{"dnsseed.kas.pa"},

	// DAG parameters
	GenesisBlock:                   &genesisBlock,
	GenesisHash:                    &genesisHash,
	PowMax:                         mainPowMax,
	BlockCoinbaseMaturity:          100,
	SubsidyReductionInterval:       210000,
	TargetTimePerBlock:             targetTimePerBlock,
	FinalityInterval:               finalityInterval,
	DifficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:    timestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1916, // 95% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016, //
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601000, // January 1, 2008 UTC
			ExpireTime: 1230767999000, // December 31, 2008 UTC
		},
	},

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
}

// RegressionNetParams defines the network parameters for the regression test
// Kaspa network. Not to be confused with the test Kaspa network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = Params{
	K:           ghostdagK,
	Name:        "regtest",
	Net:         wire.Regtest,
	RPCPort:     "16210",
	DefaultPort: "16211",
	DNSSeeds:    []string{},

	// DAG parameters
	GenesisBlock:                   &regtestGenesisBlock,
	GenesisHash:                    &regtestGenesisHash,
	PowMax:                         regressionPowMax,
	BlockCoinbaseMaturity:          100,
	SubsidyReductionInterval:       150,
	TargetTimePerBlock:             targetTimePerBlock,
	FinalityInterval:               finalityInterval,
	DifficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:    timestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 108, // 75%  of MinerConfirmationWindow
	MinerConfirmationWindow:       144,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires
		},
	},

	// Mempool parameters
	RelayNonStdTxs: true,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixKaspaReg,

	// Address encoding magics
	PrivateKeyID: 0xef, // starts with 9 (uncompressed) or c (compressed)

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	EnableNonNativeSubnetworks: false,
}

// TestnetParams defines the network parameters for the test Kaspa network.
var TestnetParams = Params{
	K:           ghostdagK,
	Name:        "testnet",
	Net:         wire.Testnet,
	RPCPort:     "16210",
	DefaultPort: "16211",
	DNSSeeds:    []string{"testnet-dnsseed.kas.pa"},

	// DAG parameters
	GenesisBlock:                   &testnetGenesisBlock,
	GenesisHash:                    &testnetGenesisHash,
	PowMax:                         testnetPowMax,
	BlockCoinbaseMaturity:          100,
	SubsidyReductionInterval:       210000,
	TargetTimePerBlock:             targetTimePerBlock,
	FinalityInterval:               finalityInterval,
	DifficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:    timestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601000, // January 1, 2008 UTC
			ExpireTime: 1230767999000, // December 31, 2008 UTC
		},
	},

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
}

// SimnetParams defines the network parameters for the simulation test Kaspa
// network. This network is similar to the normal test network except it is
// intended for private use within a group of individuals doing simulation
// testing. The functionality is intended to differ in that the only nodes
// which are specifically specified are used to create the network rather than
// following normal discovery rules. This is important as otherwise it would
// just turn into another public testnet.
var SimnetParams = Params{
	K:           ghostdagK,
	Name:        "simnet",
	Net:         wire.Simnet,
	RPCPort:     "16510",
	DefaultPort: "16511",
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                   &simnetGenesisBlock,
	GenesisHash:                    &simnetGenesisHash,
	PowMax:                         simnetPowMax,
	BlockCoinbaseMaturity:          100,
	SubsidyReductionInterval:       210000,
	TargetTimePerBlock:             targetTimePerBlock,
	FinalityInterval:               finalityInterval,
	DifficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:    timestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 75, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       100,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires
		},
	},

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
}

// DevnetParams defines the network parameters for the development Kaspa network.
var DevnetParams = Params{
	K:           ghostdagK,
	Name:        "devnet",
	Net:         wire.Devnet,
	RPCPort:     "16610",
	DefaultPort: "16611",
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                   &devnetGenesisBlock,
	GenesisHash:                    &devnetGenesisHash,
	PowMax:                         devnetPowMax,
	BlockCoinbaseMaturity:          100,
	SubsidyReductionInterval:       210000,
	TargetTimePerBlock:             targetTimePerBlock,
	FinalityInterval:               finalityInterval,
	DifficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
	TimestampDeviationTolerance:    timestampDeviationTolerance,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601000, // January 1, 2008 UTC
			ExpireTime: 1230767999000, // December 31, 2008 UTC
		},
	},

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
}

var (
	// ErrDuplicateNet describes an error where the parameters for a Kaspa
	// network could not be set due to the network already being a standard
	// network or previously-registered into this package.
	ErrDuplicateNet = errors.New("duplicate Kaspa network")
)

var (
	registeredNets = make(map[wire.KaspaNet]struct{})
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

// newHashFromStr converts the passed big-endian hex string into a
// daghash.Hash. It only differs from the one available in daghash in that
// it panics on an error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *daghash.Hash {
	hash, err := daghash.NewHashFromStr(hexStr)
	if err != nil {
		// Ordinarily I don't like panics in library code since it
		// can take applications down without them having a chance to
		// recover which is extremely annoying, however an exception is
		// being made in this case because the only way this can panic
		// is if there is an error in the hard-coded hashes. Thus it
		// will only ever potentially panic on init and therefore is
		// 100% predictable.
		panic(err)
	}
	return hash
}

func init() {
	// Register all default networks when the package is initialized.
	mustRegister(&MainnetParams)
	mustRegister(&TestnetParams)
	mustRegister(&RegressionNetParams)
	mustRegister(&SimnetParams)
}
