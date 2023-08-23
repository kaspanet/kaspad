// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dagconfig

import (
	"math/big"
	"time"

	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"

	"github.com/c4ei/yunseokyeol/app/appmessage"
	"github.com/c4ei/yunseokyeol/util/network"

	"github.com/pkg/errors"

	"github.com/c4ei/yunseokyeol/util"
)

// These variables are the DAG proof-of-work limit parameters for each default
// network.
var (
	// bigOne is 1 represented as a big.Int. It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// mainPowMax is the highest proof of work value a C4ex block can
	// have for the main network. It is the value 2^255 - 1.
	mainPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// testnetPowMax is the highest proof of work value a C4ex block
	// can have for the test network. It is the value 2^255 - 1.
	testnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// simnetPowMax is the highest proof of work value a C4ex block
	// can have for the simulation test network. It is the value 2^255 - 1.
	simnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// devnetPowMax is the highest proof of work value a C4ex block
	// can have for the development network. It is the value
	// 2^255 - 1.
	devnetPowMax = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)
)

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType uint8

// Params defines a C4ex network by its parameters. These parameters may be
// used by C4ex applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
// Params는 매개변수로 C4ex 네트워크를 정의합니다. 이러한 매개변수는 다음과 같습니다.
// 네트워크와 주소를 구별하기 위해 C4ex 애플리케이션에서 사용됩니다.
// 그리고 다른 네트워크에서 사용하기 위한 키 중 하나의 네트워크에 대한 키입니다.
type Params struct {
	// K defines the K parameter for GHOSTDAG consensus algorithm.
	// See ghostdag.go for further details.
	K externalapi.KType

	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net appmessage.C4exNet

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
	// BlockCoinbaseMaturity는 새로 채굴되기 전에 필요한 블록 수입니다.
	// 코인을 사용할 수 있습니다.
	BlockCoinbaseMaturity uint64

	// SubsidyGenesisReward SubsidyMergeSetRewardMultiplier, and
	// SubsidyPastRewardMultiplier are part of the block subsidy equation.
	// Further details: https://hashdag.medium.com/c4ex-launch-plan-9a63f4d754a6
	// SubsidyGenesisReward SubsidyMergeSetRewardMultiplier 및
	// SubsidyPastRewardMultiplier는 블록 보조금 방정식의 일부입니다.
	// 자세한 내용: https://hashdag.medium.com/c4ex-launch-plan-9a63f4d754a6
	SubsidyGenesisReward            uint64
	PreDeflationaryPhaseBaseSubsidy uint64
	DeflationaryPhaseBaseSubsidy    uint64

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// FinalityDuration is the duration of the finality window.
	FinalityDuration time.Duration

	// TimestampDeviationTolerance is the maximum offset a block timestamp
	// is allowed to be in the future before it gets delayed
	// TimestampDeviationTolerance는 블록 타임스탬프의 최대 오프셋입니다.
	// 지연되기 전에 미래에 있을 수 있습니다.
	TimestampDeviationTolerance int

	// DifficultyAdjustmentWindowSize is the size of window that is inspected
	// to calculate the required difficulty of each block.
	// TimestampDeviationTolerance는 블록 타임스탬프의 최대 예외입니다.
	// 지연되기 전에 미래에 있을 수 있습니다.
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
	// 이 필드는 다음과 같이 합의 규칙 변경에 대한 투표와 관련됩니다.
	// BIP0009에 의해 정의됩니다.
	// RuleChangeActivationThreshold는 임계값의 블록 수입니다.
	// 규칙 변경에 대해 긍정적인 투표를 한 상태 재타겟 창 상태
	// 규칙 변경을 잠그려면 캐스팅해야 합니다. 일반적으로
	// 메인 네트워크의 경우 95%, 테스트 네트워크의 경우 75%입니다.
	// MinerConfirmationWindow는 각 임계값의 블록 수입니다.
	// 상태 변경 창.
	// 배포는 투표할 특정 합의 규칙 변경 사항을 정의합니다.
	// 에.
	RuleChangeActivationThreshold uint64
	MinerConfirmationWindow       uint64

	// Mempool parameters
	RelayNonStdTxs bool

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	// AcceptUnroutable은 이 네트워크가 라우팅 불가를 허용하는지 여부를 지정합니다.
	// IP 주소(예: 10.0.0.0/8)
	AcceptUnroutable bool

	// Human-readable prefix for Bech32 encoded addresses
	// Bech32로 인코딩된 주소에 대한 사람이 읽을 수 있는 접두사
	Prefix util.Bech32Prefix

	// Address encoding magics
	PrivateKeyID byte // First byte of a WIF private key

	// EnableNonNativeSubnetworks enables non-native/coinbase transactions
	// EnableNonNativeSubnetworks는 비네이티브/코인베이스 거래를 활성화합니다.
	EnableNonNativeSubnetworks bool

	// DisableDifficultyAdjustment determine whether to use difficulty
	// DisableDifficultyAdjustment 난이도 사용 여부를 결정합니다.
	DisableDifficultyAdjustment bool

	// SkipProofOfWork indicates whether proof of work should be checked.
	// SkipProofOfWork는 작업 증명을 확인해야 하는지 여부를 나타냅니다.
	SkipProofOfWork bool

	// MaxCoinbasePayloadLength is the maximum length in bytes allowed for a block's coinbase's payload
	// MaxCoinbasePayloadLength는 블록의 코인베이스 페이로드에 허용되는 최대 길이(바이트)입니다.
	MaxCoinbasePayloadLength uint64

	// MaxBlockMass is the maximum mass a block is allowed
	// MaxBlockMass는 블록에 허용되는 최대 질량입니다.
	MaxBlockMass uint64

	// MaxBlockParents is the maximum number of blocks a block is allowed to point to
	// MaxBlockParents는 블록이 가리킬 수 있는 최대 블록 수입니다.
	MaxBlockParents externalapi.KType

	// MassPerTxByte is the number of grams that any byte
	// adds to a transaction.
	// MassPerTxByte는 임의의 바이트가 전송하는 그램 수입니다.
	// 트랜잭션에 추가합니다.
	MassPerTxByte uint64

	// MassPerScriptPubKeyByte is the number of grams that any
	// scriptPubKey byte adds to a transaction.
	// MassPerScriptPubKeyByte는 임의의 그램 수입니다.
	// scriptPubKey 바이트가 트랜잭션에 추가됩니다.
	MassPerScriptPubKeyByte uint64

	// MassPerSigOp is the number of grams that any
	// signature operation adds to a transaction.
	// MassPerSigOp는 임의의 그램 수입니다.
	// 서명 작업이 트랜잭션에 추가됩니다.
	MassPerSigOp uint64

	// MergeSetSizeLimit is the maximum number of blocks in a block's merge set
	// MergeSetSizeLimit은 블록 병합 세트의 최대 블록 수입니다.
	MergeSetSizeLimit uint64

	// CoinbasePayloadScriptPublicKeyMaxLength is the maximum allowed script public key in the coinbase's payload
	// CoinbasePayloadScriptPublicKeyMaxLength는 코인베이스 페이로드에서 허용되는 최대 스크립트 공개 키입니다.
	CoinbasePayloadScriptPublicKeyMaxLength uint8

	// PruningProofM is the 'm' constant in the pruning proof. For more details see: https://github.com/c4ei/research/issues/3
	// PruningProofM은 가지치기 증명의 'm' 상수입니다. 자세한 내용은 https://github.com/c4ei/research/issues/3을 참조하세요.
	PruningProofM uint64

	// DeflationaryPhaseDaaScore is the DAA score after which the monetary policy switches
	// to its deflationary phase
	// DeflationaryPhaseDaaScore는 통화 정책이 전환된 이후의 DAA 점수입니다.
	// 디플레이션 단계로
	DeflationaryPhaseDaaScore uint64

	DisallowDirectBlocksOnTopOfGenesis bool

	// MaxBlockLevel is the maximum possible block level.
	// MaxBlockLevel은 가능한 최대 블록 레벨입니다.
	MaxBlockLevel int

	MergeDepth uint64
}

// NormalizeRPCServerAddress returns addr with the current network default
// port appended if there is not already a port specified.
// NormalizeRPCServerAddress는 현재 네트워크 기본값으로 addr을 반환합니다.
// 포트가 아직 지정되지 않은 경우 포트가 추가됩니다.
func (p *Params) NormalizeRPCServerAddress(addr string) (string, error) {
	return network.NormalizeAddress(addr, p.RPCPort)
}

// FinalityDepth returns the finality duration represented in blocks
// FinalityDepth는 블록으로 표현된 최종 지속 기간을 반환합니다.
func (p *Params) FinalityDepth() uint64 {
	return uint64(p.FinalityDuration / p.TargetTimePerBlock)
}

// PruningDepth returns the pruning duration represented in blocks
// PruningDepth는 블록으로 표시된 가지치기 기간을 반환합니다.
func (p *Params) PruningDepth() uint64 {
	return 2*p.FinalityDepth() + 4*p.MergeSetSizeLimit*uint64(p.K) + 2*uint64(p.K) + 2
}

// MainnetParams defines the network parameters for the main C4ex network.
// MainnetParams는 기본 C4ex 네트워크에 대한 네트워크 매개변수를 정의합니다.
var MainnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "c4ex-mainnet",
	Net:         appmessage.Mainnet,
	RPCPort:     "21000", // 16110 --> 21000
	DefaultPort: "21001", // 16111 --> 21001
	DNSSeeds: []string{
		// This DNS seeder is run by Wolfie
		"dnsseed.c4ex.net",
		// "mainnet-dnsseed.kas.pa",
		// This DNS seeder is run by Denis Mashkevich
		// "mainnet-dnsseed-1.c4exnet.org",
		// // This DNS seeder is run by Denis Mashkevich
		// "mainnet-dnsseed-2.c4exnet.org",
		// // This DNS seeder is run by Constantine Bytensky
		// "dnsseed.cbytensky.org",
		// // This DNS seeder is run by Georges Künzli
		// "seeder1.c4exd.net",
		// // This DNS seeder is run by Georges Künzli
		// "seeder2.c4exd.net",
		// // This DNS seeder is run by Georges Künzli
		// "seeder3.c4exd.net",
		// // This DNS seeder is run by Georges Künzli
		// "seeder4.c4exd.net",
		// // This DNS seeder is run by Tim
		// "c4exdns.c4excalc.net",
	},

	// DAG parameters
	GenesisBlock:                    &genesisBlock,
	GenesisHash:                     genesisHash,
	PowMax:                          mainPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
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
	Prefix: util.Bech32PrefixC4ex,

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
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,
	DisallowDirectBlocksOnTopOfGenesis:      true,

	// This is technically 255, but we clamped it at 256 - block level of mainnet genesis
	// This means that any block that has a level lower or equal to genesis will be level 0.
	MaxBlockLevel: 225,
	MergeDepth:    defaultMergeDepth,
}

// TestnetParams defines the network parameters for the test C4ex network.
var TestnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "c4ex-testnet", // c4ex-testnet-10
	Net:         appmessage.Testnet,
	RPCPort:     "22000", // 16210 --> 22000
	DefaultPort: "22001", // 16211 --> 22001
	DNSSeeds: []string{
		"test-dnsseed.c4ex.net",
		// "testnet-10-dnsseed.kas.pa",
		// This DNS seeder is run by Tiram
		// "seeder1-testnet.c4exd.net",
	},

	// DAG parameters
	GenesisBlock:                    &testnetGenesisBlock,
	GenesisHash:                     testnetGenesisHash,
	PowMax:                          testnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
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
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixC4exTest,

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
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,

	MaxBlockLevel: 250,
	MergeDepth:    defaultMergeDepth,
}

// SimnetParams defines the network parameters for the simulation test C4ex
// network. This network is similar to the normal test network except it is
// intended for private use within a group of individuals doing simulation
// testing. The functionality is intended to differ in that the only nodes
// which are specifically specified are used to create the network rather than
// following normal discovery rules. This is important as otherwise it would
// just turn into another public testnet.
// SimnetParams는 C4ex 시뮬레이션 테스트를 위한 네트워크 매개변수를 정의합니다.
// 네트워크. 이 네트워크는 다음을 제외하면 일반 테스트 네트워크와 유사합니다.
// 시뮬레이션을 수행하는 개인 그룹 내에서 개인적으로 사용하기 위한 것입니다.
// 테스트 중입니다. 기능은 유일한 노드라는 점에서 다릅니다.
// 특별히 지정된 것은 네트워크를 생성하는 데 사용됩니다.
// 일반적인 검색 규칙을 따릅니다. 그렇지 않은 경우에는 이것이 중요합니다.
// 다른 공개 테스트넷으로 전환합니다.
var SimnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "c4ex-simnet",
	Net:         appmessage.Simnet,
	RPCPort:     "22510",    // 16510 --> 22510
	DefaultPort: "22511",    // 16511 --> 22511
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                    &simnetGenesisBlock,
	GenesisHash:                     simnetGenesisHash,
	PowMax:                          simnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
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
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: false,

	PrivateKeyID: 0x64, // starts with 4 (uncompressed) or F (compressed)
	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixC4exSim,

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
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,

	MaxBlockLevel: 250,
	MergeDepth:    defaultMergeDepth,
}

// DevnetParams defines the network parameters for the development C4ex network.
var DevnetParams = Params{
	K:           defaultGHOSTDAGK,
	Name:        "c4ex-devnet",
	Net:         appmessage.Devnet,
	RPCPort:     "22610",    // 16610 --> 22610
	DefaultPort: "22611",    // 16611 --> 22611
	DNSSeeds:    []string{}, // NOTE: There must NOT be any seeds.

	// DAG parameters
	GenesisBlock:                    &devnetGenesisBlock,
	GenesisHash:                     devnetGenesisHash,
	PowMax:                          devnetPowMax,
	BlockCoinbaseMaturity:           100,
	SubsidyGenesisReward:            defaultSubsidyGenesisReward,
	PreDeflationaryPhaseBaseSubsidy: defaultPreDeflationaryPhaseBaseSubsidy,
	DeflationaryPhaseBaseSubsidy:    defaultDeflationaryPhaseBaseSubsidy,
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
	RelayNonStdTxs: false,

	// AcceptUnroutable specifies whether this network accepts unroutable
	// IP addresses, such as 10.0.0.0/8
	AcceptUnroutable: true,

	// Human-readable part for Bech32 encoded addresses
	Prefix: util.Bech32PrefixC4exDev,

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
	DeflationaryPhaseDaaScore:               defaultDeflationaryPhaseDaaScore,

	MaxBlockLevel: 250,
	MergeDepth:    defaultMergeDepth,
}

// ErrDuplicateNet describes an error where the parameters for a C4ex
// network could not be set due to the network already being a standard
// network or previously-registered into this package.
var ErrDuplicateNet = errors.New("duplicate C4ex network")

var registeredNets = make(map[appmessage.C4exNet]struct{})

// Register registers the network parameters for a C4ex network. This may
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
