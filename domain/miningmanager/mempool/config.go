package mempool

import (
	"time"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

const (
	defaultTransactionExpireIntervalSeconds     uint64 = 60
	defaultTransactionExpireScanIntervalSeconds uint64 = 10
	defaultOrphanExpireIntervalSeconds          uint64 = 60
	defaultOrphanExpireScanIntervalSeconds      uint64 = 10

	defaultMaximumOrphanTransactionSize = 100000
	defaultMaximumOrphanTransactions    = 50
)

type config struct {
	transactionExpireIntervalDAAScore     uint64
	transactionExpireScanIntervalDAAScore uint64
	orphanExpireIntervalDAAScore          uint64
	orphanExpireScanIntervalDAAScore      uint64
	maximumOrphanTransactionSize          int
}

func defaultConfig(dagParams *dagconfig.Params) *config {
	targetBlocksPerSecond := uint64(dagParams.TargetTimePerBlock / time.Second)

	return &config{
		transactionExpireIntervalDAAScore:     defaultTransactionExpireIntervalSeconds / targetBlocksPerSecond,
		transactionExpireScanIntervalDAAScore: defaultTransactionExpireScanIntervalSeconds / targetBlocksPerSecond,
		orphanExpireIntervalDAAScore:          defaultOrphanExpireIntervalSeconds / targetBlocksPerSecond,
		orphanExpireScanIntervalDAAScore:      defaultOrphanExpireScanIntervalSeconds / targetBlocksPerSecond,
		maximumOrphanTransactionSize:          defaultMaximumOrphanTransactionSize,
	}
}
