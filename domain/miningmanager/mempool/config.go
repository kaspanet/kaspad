package mempool

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

const (
	defaultMaximumTransactionCount = 1_000_000

	defaultTransactionExpireIntervalSeconds     uint64 = 60
	defaultTransactionExpireScanIntervalSeconds uint64 = 10
	defaultOrphanExpireIntervalSeconds          uint64 = 60
	defaultOrphanExpireScanIntervalSeconds      uint64 = 10

	defaultMaximumOrphanTransactionMass = 100000
	// defaultMaximumOrphanTransactionCount should remain small as long as we have recursion in
	// removeOrphans when removeRedeemers = true
	defaultMaximumOrphanTransactionCount = 50

	// defaultMinimumRelayTransactionFee specifies the minimum transaction fee for a transaction to be accepted to
	// the mempool and relayed. It is specified in sompi per 1kg (or 1000 grams) of transaction mass.
	defaultMinimumRelayTransactionFee = util.Amount(1000)

	// Standard transaction version range might be different from what consensus accepts, therefore
	// we define separate values in mempool.
	// However, currently there's exactly one transaction version, so mempool accepts the same version
	// as consensus.
	defaultMinimumStandardTransactionVersion = constants.MaxTransactionVersion
	defaultMaximumStandardTransactionVersion = constants.MaxTransactionVersion
)

// Config represents a mempool configuration
type Config struct {
	MaximumTransactionCount               uint64
	TransactionExpireIntervalDAAScore     uint64
	TransactionExpireScanIntervalDAAScore uint64
	TransactionExpireScanIntervalSeconds  uint64
	OrphanExpireIntervalDAAScore          uint64
	OrphanExpireScanIntervalDAAScore      uint64
	MaximumOrphanTransactionMass          uint64
	MaximumOrphanTransactionCount         uint64
	AcceptNonStandard                     bool
	MaximumMassPerBlock                   uint64
	MinimumRelayTransactionFee            util.Amount
	MinimumStandardTransactionVersion     uint16
	MaximumStandardTransactionVersion     uint16
}

// DefaultConfig returns the default mempool configuration
func DefaultConfig(dagParams *dagconfig.Params) *Config {
	targetBlocksPerSecond := time.Second.Seconds() / dagParams.TargetTimePerBlock.Seconds()

	return &Config{
		MaximumTransactionCount:               defaultMaximumTransactionCount,
		TransactionExpireIntervalDAAScore:     uint64(float64(defaultTransactionExpireIntervalSeconds) / targetBlocksPerSecond),
		TransactionExpireScanIntervalDAAScore: uint64(float64(defaultTransactionExpireScanIntervalSeconds) / targetBlocksPerSecond),
		TransactionExpireScanIntervalSeconds:  defaultTransactionExpireScanIntervalSeconds,
		OrphanExpireIntervalDAAScore:          uint64(float64(defaultOrphanExpireIntervalSeconds) / targetBlocksPerSecond),
		OrphanExpireScanIntervalDAAScore:      uint64(float64(defaultOrphanExpireScanIntervalSeconds) / targetBlocksPerSecond),
		MaximumOrphanTransactionMass:          defaultMaximumOrphanTransactionMass,
		MaximumOrphanTransactionCount:         defaultMaximumOrphanTransactionCount,
		AcceptNonStandard:                     dagParams.RelayNonStdTxs,
		MaximumMassPerBlock:                   dagParams.MaxBlockMass,
		MinimumRelayTransactionFee:            defaultMinimumRelayTransactionFee,
		MinimumStandardTransactionVersion:     defaultMinimumStandardTransactionVersion,
		MaximumStandardTransactionVersion:     defaultMaximumStandardTransactionVersion,
	}
}
