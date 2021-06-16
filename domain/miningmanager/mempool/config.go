package mempool

import (
	"time"

	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

const (
	defaultMaximumTransactionCount = 1_000_000

	defaultTransactionExpireIntervalSeconds     uint64 = 60
	defaultTransactionExpireScanIntervalSeconds uint64 = 10
	defaultOrphanExpireIntervalSeconds          uint64 = 60
	defaultOrphanExpireScanIntervalSeconds      uint64 = 10

	defaultMaximumOrphanTransactionSize  = 100000
	defaultMaximumOrphanTransactionCount = 50

	defaultMinimumRelayFee = util.Amount(1000)
)

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
}

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
		MinimumRelayTransactionFee:            minimumRelayTransactionFee,
	}
}
