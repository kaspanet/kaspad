package multisetstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/util/staging"
)

var bucketName = []byte("multisets")

// multisetStore represents a store of Multisets
type multisetStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New instantiates a new MultisetStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.MultisetStore {
	return &multisetStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

// Stage stages the given multiset for the given blockHash
func (ms *multisetStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, multiset model.Multiset) {
	stagingShard := ms.stagingShard(stagingArea)

	stagingShard.toAdd[*blockHash] = multiset.Clone()
}

func (ms *multisetStore) IsStaged(stagingArea *model.StagingArea) bool {
	return ms.stagingShard(stagingArea).isStaged()
}

// Get gets the multiset associated with the given blockHash
func (ms *multisetStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (model.Multiset, error) {
	stagingShard := ms.stagingShard(stagingArea)

	if multiset, ok := stagingShard.toAdd[*blockHash]; ok {
		return multiset.Clone(), nil
	}

	if multiset, ok := ms.cache.Get(blockHash); ok {
		return multiset.(model.Multiset).Clone(), nil
	}

	multisetBytes, err := dbContext.Get(ms.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	multiset, err := ms.deserializeMultiset(multisetBytes)
	if err != nil {
		return nil, err
	}
	ms.cache.Add(blockHash, multiset)
	return multiset.Clone(), nil
}

// Delete deletes the multiset associated with the given blockHash
func (ms *multisetStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	stagingShard := ms.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		delete(stagingShard.toAdd, *blockHash)
		return
	}
	stagingShard.toDelete[*blockHash] = struct{}{}
}

func (ms *multisetStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return ms.bucket.Key(hash.ByteSlice())
}

func (ms *multisetStore) serializeMultiset(multiset model.Multiset) ([]byte, error) {
	return proto.Marshal(serialization.MultisetToDBMultiset(multiset))
}

func (ms *multisetStore) deserializeMultiset(multisetBytes []byte) (model.Multiset, error) {
	dbMultiset := &serialization.DbMultiset{}
	err := proto.Unmarshal(multisetBytes, dbMultiset)
	if err != nil {
		return nil, err
	}

	return serialization.DBMultisetToMultiset(dbMultiset)
}
