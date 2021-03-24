package model

import "github.com/pkg/errors"

type StagingShard interface {
	Commit(dbTx DBTransaction) error
}

type StagingArea struct {
	shards     map[string]StagingShard
	isCommited bool
}

func NewStagingArea() *StagingArea {
	return &StagingArea{
		shards:     map[string]StagingShard{},
		isCommited: false,
	}
}

func (sa *StagingArea) GetOrCreateShard(shardName string, createFunc func() StagingShard) StagingShard {
	if _, ok := sa.shards[shardName]; !ok {
		sa.shards[shardName] = createFunc()
	}

	return sa.shards[shardName]
}

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
