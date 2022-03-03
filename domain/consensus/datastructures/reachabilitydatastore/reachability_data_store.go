package reachabilitydatastore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
)

var reachabilityDataBucketName = []byte("reachability-data")
var reachabilityReindexRootKeyName = []byte("reachability-reindex-root")

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
	shardID                      model.StagingShardID
	reachabilityDataCache        *lrucache.LRUCache
	reachabilityReindexRootCache *externalapi.DomainHash

	reachabilityDataBucket     model.DBBucket
	reachabilityReindexRootKey model.DBKey
}

// New instantiates a new ReachabilityDataStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.ReachabilityDataStore {
	return &reachabilityDataStore{
		shardID:                    staging.GenerateShardingID(),
		reachabilityDataCache:      lrucache.New(cacheSize, preallocate),
		reachabilityDataBucket:     prefixBucket.Bucket(reachabilityDataBucketName),
		reachabilityReindexRootKey: prefixBucket.Key(reachabilityReindexRootKeyName),
	}
}

// StageReachabilityData stages the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) StageReachabilityData(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, reachabilityData model.ReachabilityData) {
	stagingShard := rds.stagingShard(stagingArea)

	stagingShard.reachabilityData[*blockHash] = reachabilityData
}

func (rds *reachabilityDataStore) Delete(dbContext model.DBWriter) error {
	cursor, err := dbContext.Cursor(rds.reachabilityDataBucket)
	if err != nil {
		return err
	}

	for ok := cursor.First(); ok; ok = cursor.Next() {
		key, err := cursor.Key()
		if err != nil {
			return err
		}

		err = dbContext.Delete(key)
		if err != nil {
			return err
		}
	}

	return dbContext.Delete(rds.reachabilityReindexRootKey)
}

// StageReachabilityReindexRoot stages the given reachabilityReindexRoot
func (rds *reachabilityDataStore) StageReachabilityReindexRoot(stagingArea *model.StagingArea, reachabilityReindexRoot *externalapi.DomainHash) {
	stagingShard := rds.stagingShard(stagingArea)

	stagingShard.reachabilityReindexRoot = reachabilityReindexRoot
}

func (rds *reachabilityDataStore) IsStaged(stagingArea *model.StagingArea) bool {
	return rds.stagingShard(stagingArea).isStaged()
}

var errNotFound = errors.Wrap(database.ErrNotFound, "reachability data not found")

// ReachabilityData returns the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) ReachabilityData(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (model.ReachabilityData, error) {
	stagingShard := rds.stagingShard(stagingArea)

	if reachabilityData, ok := stagingShard.reachabilityData[*blockHash]; ok {
		return reachabilityData, nil
	}

	if reachabilityData, ok := rds.reachabilityDataCache.Get(blockHash); ok {
		if reachabilityData == nil {
			return nil, errNotFound
		}
		return reachabilityData.(model.ReachabilityData), nil
	}

	reachabilityDataBytes, err := dbContext.Get(rds.reachabilityDataBlockHashAsKey(blockHash))
	if database.IsNotFoundError(err) {
		rds.reachabilityDataCache.Add(blockHash, nil)
	}
	if err != nil {
		return nil, err
	}

	reachabilityData, err := rds.deserializeReachabilityData(reachabilityDataBytes)
	if err != nil {
		return nil, err
	}
	rds.reachabilityDataCache.Add(blockHash, reachabilityData)
	return reachabilityData, nil
}

func (rds *reachabilityDataStore) HasReachabilityData(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	_, err := rds.ReachabilityData(dbContext, stagingArea, blockHash)
	if database.IsNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// ReachabilityReindexRoot returns the current reachability reindex root
func (rds *reachabilityDataStore) ReachabilityReindexRoot(dbContext model.DBReader, stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	stagingShard := rds.stagingShard(stagingArea)

	if stagingShard.reachabilityReindexRoot != nil {
		return stagingShard.reachabilityReindexRoot, nil
	}

	if rds.reachabilityReindexRootCache != nil {
		return rds.reachabilityReindexRootCache, nil
	}

	reachabilityReindexRootBytes, err := dbContext.Get(rds.reachabilityReindexRootKey)
	if err != nil {
		return nil, err
	}

	reachabilityReindexRoot, err := rds.deserializeReachabilityReindexRoot(reachabilityReindexRootBytes)
	if err != nil {
		return nil, err
	}
	rds.reachabilityReindexRootCache = reachabilityReindexRoot
	return reachabilityReindexRoot, nil
}

func (rds *reachabilityDataStore) reachabilityDataBlockHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return rds.reachabilityDataBucket.Key(hash.ByteSlice())
}

func (rds *reachabilityDataStore) serializeReachabilityData(reachabilityData model.ReachabilityData) ([]byte, error) {
	return proto.Marshal(serialization.ReachablityDataToDBReachablityData(reachabilityData))
}

func (rds *reachabilityDataStore) deserializeReachabilityData(reachabilityDataBytes []byte) (model.ReachabilityData, error) {
	dbReachabilityData := &serialization.DbReachabilityData{}
	err := proto.Unmarshal(reachabilityDataBytes, dbReachabilityData)
	if err != nil {
		return nil, err
	}

	return serialization.DBReachablityDataToReachablityData(dbReachabilityData)
}

func (rds *reachabilityDataStore) serializeReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash) ([]byte, error) {
	return proto.Marshal(serialization.DomainHashToDbHash(reachabilityReindexRoot))
}

func (rds *reachabilityDataStore) deserializeReachabilityReindexRoot(reachabilityReindexRootBytes []byte) (*externalapi.DomainHash, error) {
	dbHash := &serialization.DbHash{}
	err := proto.Unmarshal(reachabilityReindexRootBytes, dbHash)
	if err != nil {
		return nil, err
	}

	return serialization.DbHashToDomainHash(dbHash)
}
