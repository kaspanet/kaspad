package model

type StagingShard interface {
	Commit(dbTx DBTransaction) error
}

type StagingArea struct {
	shards map[string]StagingShard
}

func (sa *StagingArea) GetOrCreateShard(shardName string, createFunc func() StagingShard) StagingShard {
	if _, ok := sa.shards[shardName]; !ok {
		sa.shards[shardName] = createFunc()
	}

	return sa.shards[shardName]
}

func (sa *StagingArea) Commit(dbTx DBTransaction) error {
	for _, shard := range sa.shards {
		err := shard.Commit(dbTx)
		if err != nil {
			return err
		}
	}
	return nil
}
