package model

import "github.com/pkg/errors"

// StagingShard is an interface that enables every store to have it's own Commit logic
// See StagingArea for more details
type StagingShard interface {
	Commit(dbTx DBTransaction) error
}

// StagingArea is single changeset inside the consensus database, similar to a transaction in a classic database.
// Each StagingArea consists of multiple StagingShards, one for each dataStore that has any changes within it.
// To enable maximum flexibility for all stores, each has to define it's own Commit method, and pass it to the
// StagingArea through the relevant StagingShard.
//
// When the StagingArea is being Committed, it goes over all it's shards, and commits those one-by-one.
// Since Commit happens in a DatabaseTransaction, a StagingArea is atomic.
type StagingArea struct {
	shards     map[string]StagingShard
	isCommited bool
}

// NewStagingArea creates a new, empty staging area.
func NewStagingArea() *StagingArea {
	return &StagingArea{
		shards:     map[string]StagingShard{},
		isCommited: false,
	}
}

// GetOrCreateShard attempts to retrieve a shard with the given name.
// If it does not exist - a new shard is created using `createFunc`.
func (sa *StagingArea) GetOrCreateShard(shardName string, createFunc func() StagingShard) StagingShard {
	if _, ok := sa.shards[shardName]; !ok {
		sa.shards[shardName] = createFunc()
	}

	return sa.shards[shardName]
}

// Commit goes over all the Shards in the StagingArea and commits them, inside the provided database transaction.
// Note: the transaction itself is not committed, this is the callers responsibility to commit it.
func (sa *StagingArea) Commit(dbTx DBTransaction) error {
	if sa.isCommited {
		return errors.New("Attempt to call Commit on already committed stagingArea")
	}

	for _, shard := range sa.shards {
		err := shard.Commit(dbTx)
		if err != nil {
			return err
		}
	}

	sa.isCommited = true

	return nil
}
