package staging

import (
	"sync/atomic"

	"github.com/c4ei/YunSeokYeol/domain/consensus/model"
	"github.com/c4ei/YunSeokYeol/infrastructure/logger"
)

// CommitAllChanges creates a transaction in `databaseContext`, and commits all changes in `stagingArea` through it.
func CommitAllChanges(databaseContext model.DBManager, stagingArea *model.StagingArea) error {
	onEnd := logger.LogAndMeasureExecutionTime(utilLog, "commitAllChanges")
	defer onEnd()

	dbTx, err := databaseContext.Begin()
	if err != nil {
		return err
	}

	err = stagingArea.Commit(dbTx)
	if err != nil {
		return err
	}

	return dbTx.Commit()
}

var lastShardingID uint64

// GenerateShardingID generates a unique staging sharding ID.
func GenerateShardingID() model.StagingShardID {
	return model.StagingShardID(atomic.AddUint64(&lastShardingID, 1))
}
