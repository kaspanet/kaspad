package subsidystore

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/util/staging"
)

var bucketName = []byte("subsidies")

type subsidyStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New instantiates a new SubsidyStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.SubsidyStore {
	return &subsidyStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

func (fs *subsidyStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, subsidy uint64) {
	stagingShard := fs.stagingShard(stagingArea)

	stagingShard.toAdd[*blockHash] = subsidy
}

func (fs *subsidyStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	stagingShard := fs.stagingShard(stagingArea)

	if subsidy, ok := stagingShard.toAdd[*blockHash]; ok {
		return subsidy, nil
	}

	if subsidy, ok := fs.cache.Get(blockHash); ok {
		return subsidy.(uint64), nil
	}

	subsidyBytes, err := dbContext.Get(fs.hashAsKey(blockHash))
	if err != nil {
		return 0, err
	}
	subsidy := fs.deserializeSubsidy(subsidyBytes)

	fs.cache.Add(blockHash, subsidy)
	return subsidy, nil
}

func (fs *subsidyStore) Has(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	panic("implement me")
}

func (fs *subsidyStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	panic("implement me")
}

func (fs *subsidyStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return fs.bucket.Key(hash.ByteSlice())
}

func (fs *subsidyStore) serializeSubsidy(subsidy uint64) []byte {
	var subsidyBytes [8]byte
	binary.BigEndian.PutUint64(subsidyBytes[:], subsidy)
	return subsidyBytes[:]
}

func (fs *subsidyStore) deserializeSubsidy(subsidyBytes []byte) uint64 {
	return binary.BigEndian.Uint64(subsidyBytes)
}
