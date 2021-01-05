package reachabilitydatastore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
)

var reachabilityDataBucket = dbkeys.MakeBucket([]byte("reachability-data"))
var reachabilityIntervalBucket = dbkeys.MakeBucket([]byte("reachability-interval"))
var reachabilityReindexRootKey = dbkeys.MakeBucket().Key([]byte("reachability-reindex-root"))

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
	reachabilityDataStaging        map[externalapi.DomainHash]model.ReachabilityData
	reachabilityIntervalStaging    map[externalapi.DomainHash]*model.ReachabilityInterval
	reachabilityReindexRootStaging *externalapi.DomainHash
	reachabilityDataCache          *lrucache.LRUCache
	reachabilityIntervalCache      *lrucache.LRUCache
	reachabilityReindexRootCache   *externalapi.DomainHash
}

// New instantiates a new ReachabilityDataStore
func New(cacheSize int) model.ReachabilityDataStore {
	return &reachabilityDataStore{
		reachabilityDataStaging:     make(map[externalapi.DomainHash]model.ReachabilityData),
		reachabilityIntervalStaging: make(map[externalapi.DomainHash]*model.ReachabilityInterval),
		reachabilityDataCache:       lrucache.New(cacheSize),
		reachabilityIntervalCache:   lrucache.New(cacheSize),
	}
}

func (rds *reachabilityDataStore) StageInterval(blockHash *externalapi.DomainHash, interval *model.ReachabilityInterval) {
	rds.reachabilityIntervalStaging[*blockHash] = interval.Clone()
}

func (rds *reachabilityDataStore) StageReachabilityData(blockHash *externalapi.DomainHash,
	reachabilityData model.ReachabilityData) {

	rds.reachabilityDataStaging[*blockHash] = reachabilityData
}

// StageReachabilityReindexRoot stages the given reachabilityReindexRoot
func (rds *reachabilityDataStore) StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash) {
	rds.reachabilityReindexRootStaging = reachabilityReindexRoot
}

func (rds *reachabilityDataStore) IsAnythingStaged() bool {
	return len(rds.reachabilityDataStaging) != 0 || rds.reachabilityReindexRootStaging != nil
}

func (rds *reachabilityDataStore) Discard() {
	rds.reachabilityDataStaging = make(map[externalapi.DomainHash]model.ReachabilityData)
	rds.reachabilityReindexRootStaging = nil
}

func (rds *reachabilityDataStore) Commit(dbTx model.DBTransaction) error {
	if rds.reachabilityReindexRootStaging != nil {
		reachabilityReindexRootBytes, err := rds.serializeReachabilityReindexRoot(rds.reachabilityReindexRootStaging)
		if err != nil {
			return err
		}
		err = dbTx.Put(reachabilityReindexRootKey, reachabilityReindexRootBytes)
		if err != nil {
			return err
		}
		rds.reachabilityReindexRootCache = rds.reachabilityReindexRootStaging
	}

	for hash, reachabilityData := range rds.reachabilityDataStaging {
		reachabilityDataBytes, err := rds.serializeReachabilityData(reachabilityData)
		if err != nil {
			return err
		}
		err = dbTx.Put(rds.reachabilityDataBlockHashAsKey(&hash), reachabilityDataBytes)
		if err != nil {
			return err
		}
		rds.reachabilityDataCache.Add(&hash, reachabilityData)
	}

	for hash, interval := range rds.reachabilityIntervalStaging {
		intervalBytes, err := rds.serializeInterval(interval)
		if err != nil {
			return err
		}
		err = dbTx.Put(rds.reachabilityIntervalBlockHashAsKey(&hash), intervalBytes)
		if err != nil {
			return err
		}
		rds.reachabilityIntervalCache.Add(&hash, interval)
	}

	rds.Discard()
	return nil
}

// ReachabilityData returns the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) ReachabilityData(dbContext model.DBReader,
	blockHash *externalapi.DomainHash) (model.ReachabilityData, error) {

	if reachabilityData, ok := rds.reachabilityDataStaging[*blockHash]; ok {
		return reachabilityData, nil
	}

	if reachabilityData, ok := rds.reachabilityDataCache.Get(blockHash); ok {
		return reachabilityData.(model.ReachabilityData), nil
	}

	reachabilityDataBytes, err := dbContext.Get(rds.reachabilityDataBlockHashAsKey(blockHash))
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

func (rds *reachabilityDataStore) Interval(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	if interval, ok := rds.reachabilityIntervalStaging[*blockHash]; ok {
		return interval, nil
	}

	if interval, ok := rds.reachabilityIntervalCache.Get(blockHash); ok {
		return interval.(*model.ReachabilityInterval), nil
	}

	intervalBytes, err := dbContext.Get(rds.reachabilityIntervalBlockHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	interval, err := rds.deserializeReachabilityInterval(intervalBytes)
	if err != nil {
		return nil, err
	}

	rds.reachabilityIntervalCache.Add(blockHash, interval)
	return interval, nil
}

func (rds *reachabilityDataStore) HasReachabilityData(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := rds.reachabilityDataStaging[*blockHash]; ok {
		return true, nil
	}

	if rds.reachabilityDataCache.Has(blockHash) {
		return true, nil
	}

	return dbContext.Has(rds.reachabilityDataBlockHashAsKey(blockHash))
}

// ReachabilityReindexRoot returns the current reachability reindex root
func (rds *reachabilityDataStore) ReachabilityReindexRoot(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if rds.reachabilityReindexRootStaging != nil {
		return rds.reachabilityReindexRootStaging, nil
	}

	if rds.reachabilityReindexRootCache != nil {
		return rds.reachabilityReindexRootCache, nil
	}

	reachabilityReindexRootBytes, err := dbContext.Get(reachabilityReindexRootKey)
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

func (rds *reachabilityDataStore) reachabilityIntervalBlockHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return reachabilityIntervalBucket.Key(hash.ByteSlice())
}

func (rds *reachabilityDataStore) reachabilityDataBlockHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return reachabilityDataBucket.Key(hash.ByteSlice())
}

func (rds *reachabilityDataStore) serializeInterval(interval *model.ReachabilityInterval) ([]byte, error) {
	return proto.Marshal(serialization.ReachablityIntervalToDBReachablityInterval(interval))
}

func (rds *reachabilityDataStore) deserializeReachabilityInterval(intervalBytes []byte) (*model.ReachabilityInterval, error) {
	dbReachabilityInterval := &serialization.DbReachabilityInterval{}
	err := proto.Unmarshal(intervalBytes, dbReachabilityInterval)
	if err != nil {
		return nil, err
	}

	return serialization.DBReachablityIntervalToReachablityInterval(dbReachabilityInterval), nil
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
