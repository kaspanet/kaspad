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
)

type config struct {
	transactionExpireIntervalDAAScore     uint64
	transactionExpireScanIntervalDAAScore uint64
	orphanExpireIntervalDAAScore          uint64
	orphanExpireScanIntervalDAAScore      uint64
}

func defaultConfig(dagParams *dagconfig.Params) *config {
	targetBlocksPerSecond := uint64(dagParams.TargetTimePerBlock / time.Second)

	return &config{
		transactionExpireIntervalDAAScore:     defaultTransactionExpireIntervalSeconds / targetBlocksPerSecond,
		transactionExpireScanIntervalDAAScore: defaultTransactionExpireScanIntervalSeconds / targetBlocksPerSecond,
		orphanExpireIntervalDAAScore:          defaultOrphanExpireIntervalSeconds / targetBlocksPerSecond,
		orphanExpireScanIntervalDAAScore:      defaultOrphanExpireScanIntervalSeconds / targetBlocksPerSecond,
	}
}
