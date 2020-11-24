package reachabilitydatastore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/golang-lru/simplelru"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var reachabilityDataBucket = dbkeys.MakeBucket([]byte("reachability-data"))
var reachabilityReindexRootKey = dbkeys.MakeBucket().Key([]byte("reachability-reindex-root"))

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
	reachabilityDataStaging        map[externalapi.DomainHash]*model.ReachabilityData
	reachabilityReindexRootStaging *externalapi.DomainHash
	cache                          simplelru.LRUCache
}

// New instantiates a new ReachabilityDataStore
func New(cacheSize int) (model.ReachabilityDataStore, error) {
	reachabilityDataStore := &reachabilityDataStore{
		reachabilityDataStaging:        make(map[externalapi.DomainHash]*model.ReachabilityData),
		reachabilityReindexRootStaging: nil,
	}

	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		return nil, err
	}
	reachabilityDataStore.cache = cache

	return reachabilityDataStore, nil
}

// StageReachabilityData stages the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) StageReachabilityData(blockHash *externalapi.DomainHash,
	reachabilityData *model.ReachabilityData) error {
	clone, err := rds.cloneReachabilityData(reachabilityData)
	if err != nil {
		return err
	}

	rds.reachabilityDataStaging[*blockHash] = clone
	return nil
}

// StageReachabilityReindexRoot stages the given reachabilityReindexRoot
func (rds *reachabilityDataStore) StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash) {
	rds.reachabilityReindexRootStaging = reachabilityReindexRoot
}

func (rds *reachabilityDataStore) IsAnythingStaged() bool {
	return len(rds.reachabilityDataStaging) != 0 || rds.reachabilityReindexRootStaging != nil
}

func (rds *reachabilityDataStore) Discard() {
	rds.reachabilityDataStaging = make(map[externalapi.DomainHash]*model.ReachabilityData)
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
	}

	rds.Discard()
	return nil
}

// ReachabilityData returns the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) ReachabilityData(dbContext model.DBReader,
	blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {

	if reachabilityData, ok := rds.reachabilityDataStaging[*blockHash]; ok {
		return reachabilityData, nil
	}

	reachabilityDataBytes, err := dbContext.Get(rds.reachabilityDataBlockHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return rds.deserializeReachabilityData(reachabilityDataBytes)
}

func (rds *reachabilityDataStore) HasReachabilityData(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := rds.reachabilityDataStaging[*blockHash]; ok {
		return true, nil
	}

	return dbContext.Has(rds.reachabilityDataBlockHashAsKey(blockHash))
}

// ReachabilityReindexRoot returns the current reachability reindex root
func (rds *reachabilityDataStore) ReachabilityReindexRoot(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	if rds.reachabilityReindexRootStaging != nil {
		return rds.reachabilityReindexRootStaging, nil
	}
	reachabilityReindexRootBytes, err := dbContext.Get(reachabilityReindexRootKey)
	if err != nil {
		return nil, err
	}

	reachabilityReindexRoot, err := rds.deserializeReachabilityReindexRoot(reachabilityReindexRootBytes)
	if err != nil {
		return nil, err
	}
	return reachabilityReindexRoot, nil
}

func (rds *reachabilityDataStore) reachabilityDataBlockHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return reachabilityDataBucket.Key(hash[:])
}

func (rds *reachabilityDataStore) serializeReachabilityData(reachabilityData *model.ReachabilityData) ([]byte, error) {
	return proto.Marshal(serialization.ReachablityDataToDBReachablityData(reachabilityData))
}

func (rds *reachabilityDataStore) deserializeReachabilityData(reachabilityDataBytes []byte) (*model.ReachabilityData, error) {
	dbReachabilityData := &serialization.DbReachabilityData{}
	err := proto.Unmarshal(reachabilityDataBytes, dbReachabilityData)
	if err != nil {
		return nil, err
	}

	return serialization.DBReachablityDataToReachablityData(dbReachabilityData), nil
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

func (rds *reachabilityDataStore) cloneReachabilityData(reachabilityData *model.ReachabilityData) (*model.ReachabilityData, error) {
	serialized, err := rds.serializeReachabilityData(reachabilityData)
	if err != nil {
		return nil, err
	}

	return rds.deserializeReachabilityData(serialized)
}
