package model

import "github.com/pkg/errors"

// StagingShard is an interface that enables every store to have it's own Commit logic
// See StagingArea for more details
type StagingShard interface {
	Commit(dbTx DBTransaction) error
}

// StagingShardID is used to identify each of the store's staging shards
type StagingShardID byte

// StagingShardID constants
const (
	StagingShardIDAcceptanceData StagingShardID = iota
	StagingShardIDBlockHeader
	StagingShardIDBlockRelation
	StagingShardIDBlockStatus
	StagingShardIDBlock
	StagingShardIDConsensusState
	StagingShardIDDAABlocks
	StagingShardIDFinality
	StagingShardIDGHOSTDAG
	StagingShardIDHeadersSelectedChain
	StagingShardIDHeadersSelectedTip
	StagingShardIDMultiset
	StagingShardIDPruning
	StagingShardIDReachabilityData
	StagingShardIDUTXODiff
	// Always leave StagingShardIDLen as the last constant
	StagingShardIDLen
)

// StagingArea is single changeset inside the consensus database, similar to a transaction in a classic database.
// Each StagingArea consists of multiple StagingShards, one for each dataStore that has any changes within it.
// To enable maximum flexibility for all stores, each has to define it's own Commit method, and pass it to the
// StagingArea through the relevant StagingShard.
//
// When the StagingArea is being Committed, it goes over all it's shards, and commits those one-by-one.
// Since Commit happens in a DatabaseTransaction, a StagingArea is atomic.
type StagingArea struct {
	shards      [StagingShardIDLen]StagingShard
	isCommitted bool
}

// NewStagingArea creates a new, empty staging area.
func NewStagingArea() *StagingArea {
	return &StagingArea{
		shards:      [StagingShardIDLen]StagingShard{},
		isCommitted: false,
	}
}

// GetOrCreateShard attempts to retrieve a shard with the given name.
// If it does not exist - a new shard is created using `createFunc`.
func (sa *StagingArea) GetOrCreateShard(shardID StagingShardID, createFunc func() StagingShard) StagingShard {
	if sa.shards[shardID] == nil {
		sa.shards[shardID] = createFunc()
	}

	return sa.shards[shardID]
}

// Commit goes over all the Shards in the StagingArea and commits them, inside the provided database transaction.
// Note: the transaction itself is not committed, this is the callers responsibility to commit it.
func (sa *StagingArea) Commit(dbTx DBTransaction) error {
	if sa.isCommitted {
		return errors.New("Attempt to call Commit on already committed stagingArea")
	}

	for _, shard := range sa.shards {
		if shard == nil {
			continue
		}
		err := shard.Commit(dbTx)
		if err != nil {
			return err
		}
	}

	sa.isCommitted = true

	return nil
}
