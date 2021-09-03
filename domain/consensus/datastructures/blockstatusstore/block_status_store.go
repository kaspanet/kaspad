package blockstatusstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/util/staging"
)

var bucketName = []byte("block-statuses")

// blockStatusStore represents a store of BlockStatuses
type blockStatusStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New instantiates a new BlockStatusStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.BlockStatusStore {
	return &blockStatusStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

// Stage stages the given blockStatus for the given blockHash
func (bss *blockStatusStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, blockStatus externalapi.BlockStatus) {
	stagingShard := bss.stagingShard(stagingArea)
	stagingShard.toAdd[*blockHash] = blockStatus.Clone()
}

func (bss *blockStatusStore) IsStaged(stagingArea *model.StagingArea) bool {
	return bss.stagingShard(stagingArea).isStaged()
}

// Get gets the blockStatus associated with the given blockHash
func (bss *blockStatusStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	stagingShard := bss.stagingShard(stagingArea)

	if status, ok := stagingShard.toAdd[*blockHash]; ok {
		return status, nil
	}

	if status, ok := bss.cache.Get(blockHash); ok {
		return status.(externalapi.BlockStatus), nil
	}

	statusBytes, err := dbContext.Get(bss.hashAsKey(blockHash))
	if err != nil {
		return 0, err
	}

	status, err := bss.deserializeBlockStatus(statusBytes)
	if err != nil {
		return 0, err
	}
	bss.cache.Add(blockHash, status)
	return status, nil
}

// Exists returns true if the blockStatus for the given blockHash exists
func (bss *blockStatusStore) Exists(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	stagingShard := bss.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		return true, nil
	}

	if bss.cache.Has(blockHash) {
		return true, nil
	}

	exists, err := dbContext.Has(bss.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (bss *blockStatusStore) serializeBlockStatus(status externalapi.BlockStatus) ([]byte, error) {
	dbBlockStatus := serialization.DomainBlockStatusToDbBlockStatus(status)
	return proto.Marshal(dbBlockStatus)
}

func (bss *blockStatusStore) deserializeBlockStatus(statusBytes []byte) (externalapi.BlockStatus, error) {
	dbBlockStatus := &serialization.DbBlockStatus{}
	err := proto.Unmarshal(statusBytes, dbBlockStatus)
	if err != nil {
		return 0, err
	}
	return serialization.DbBlockStatusToDomainBlockStatus(dbBlockStatus), nil
}

func (bss *blockStatusStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bss.bucket.Key(hash.ByteSlice())
}
