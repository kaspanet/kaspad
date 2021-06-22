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

	defaultMaximumOrphanTransactionSize = 100000
	// defaultMaximumOrphanTransactionCount should remain small as long as we have recursion in
	// removeOrphans when removeRedeemers = true
	defaultMaximumOrphanTransactionCount = 50

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
	MaximumTransactionCount               int
	TransactionExpireIntervalDAAScore     uint64
	TransactionExpireScanIntervalDAAScore uint64
	OrphanExpireIntervalDAAScore          uint64
	OrphanExpireScanIntervalDAAScore      uint64
	MaximumOrphanTransactionSize          int
	MaximumOrphanTransactionCount         int
	AcceptNonStandard                     bool
	MaximumMassAcceptedByBlock            uint64
	MinimumRelayTransactionFee            util.Amount
	MinimumStandardTransactionVersion     uint16
	MaximumStandardTransactionVersion     uint16
}

// DefaultConfig returns the default mempool configuration
func DefaultConfig(dagParams *dagconfig.Params) *Config {
	targetBlocksPerSecond := uint64(time.Second / dagParams.TargetTimePerBlock)

	return &Config{
		MaximumTransactionCount:               defaultMaximumTransactionCount,
		TransactionExpireIntervalDAAScore:     defaultTransactionExpireIntervalSeconds / targetBlocksPerSecond,
		TransactionExpireScanIntervalDAAScore: defaultTransactionExpireScanIntervalSeconds / targetBlocksPerSecond,
		OrphanExpireIntervalDAAScore:          defaultOrphanExpireIntervalSeconds / targetBlocksPerSecond,
		OrphanExpireScanIntervalDAAScore:      defaultOrphanExpireScanIntervalSeconds / targetBlocksPerSecond,
		MaximumOrphanTransactionSize:          defaultMaximumOrphanTransactionSize,
		MaximumOrphanTransactionCount:         defaultMaximumOrphanTransactionCount,
		AcceptNonStandard:                     dagParams.RelayNonStdTxs,
		MaximumMassAcceptedByBlock:            dagParams.MaxMassAcceptedByBlock,
		MinimumRelayTransactionFee:            defaultMinimumRelayTransactionFee,
		MinimumStandardTransactionVersion:     defaultMinimumStandardTransactionVersion,
		MaximumStandardTransactionVersion:     defaultMaximumStandardTransactionVersion,
	}
}
