package mempool

import (
	"time"

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

	defaultAcceptNonStandard = false
)

type config struct {
	maximumTransactionCount               int
	transactionExpireIntervalDAAScore     uint64
	transactionExpireScanIntervalDAAScore uint64
	orphanExpireIntervalDAAScore          uint64
	orphanExpireScanIntervalDAAScore      uint64
	maximumOrphanTransactionSize          int
	maximumOrphanTransactionCount         int
	acceptNonStandard                     bool
	maximumMassAcceptedByBlock            uint64
}

func defaultConfig(dagParams *dagconfig.Params) *config {
	targetBlocksPerSecond := uint64(dagParams.TargetTimePerBlock / time.Second)

	return &config{
		maximumTransactionCount:               defaultMaximumTransactionCount,
		transactionExpireIntervalDAAScore:     defaultTransactionExpireIntervalSeconds / targetBlocksPerSecond,
		transactionExpireScanIntervalDAAScore: defaultTransactionExpireScanIntervalSeconds / targetBlocksPerSecond,
		orphanExpireIntervalDAAScore:          defaultOrphanExpireIntervalSeconds / targetBlocksPerSecond,
		orphanExpireScanIntervalDAAScore:      defaultOrphanExpireScanIntervalSeconds / targetBlocksPerSecond,
		maximumOrphanTransactionSize:          defaultMaximumOrphanTransactionSize,
		maximumOrphanTransactionCount:         defaultMaximumOrphanTransactionCount,
		acceptNonStandard:                     defaultAcceptNonStandard,
		maximumMassAcceptedByBlock:            dagParams.MaxMassAcceptedByBlock,
	}
}
