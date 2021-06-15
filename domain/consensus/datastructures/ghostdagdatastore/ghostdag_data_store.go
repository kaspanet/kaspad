package ghostdagdatastore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
)

var bucketName = []byte("block-ghostdag-data")

// ghostdagDataStore represents a store of BlockGHOSTDAGData
type ghostdagDataStore struct {
	cache  *lrucache.LRUCache
	bucket model.DBBucket
}

// New instantiates a new GHOSTDAGDataStore
func New(prefix *prefix.Prefix, cacheSize int, preallocate bool) model.GHOSTDAGDataStore {
	return &ghostdagDataStore{
		cache:  lrucache.New(cacheSize, preallocate),
		bucket: database.MakeBucket(prefix.Serialize()).Bucket(bucketName),
	}
}

// Stage stages the given blockGHOSTDAGData for the given blockHash
func (gds *ghostdagDataStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {
	stagingShard := gds.stagingShard(stagingArea)

	stagingShard.toAdd[*blockHash] = blockGHOSTDAGData
}

func (gds *ghostdagDataStore) IsStaged(stagingArea *model.StagingArea) bool {
	return gds.stagingShard(stagingArea).isStaged()
}

// Get gets the blockGHOSTDAGData associated with the given blockHash
func (gds *ghostdagDataStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*model.BlockGHOSTDAGData, error) {
	stagingShard := gds.stagingShard(stagingArea)

	if blockGHOSTDAGData, ok := stagingShard.toAdd[*blockHash]; ok {
		return blockGHOSTDAGData, nil
	}

	if blockGHOSTDAGData, ok := gds.cache.Get(blockHash); ok {
		return blockGHOSTDAGData.(*model.BlockGHOSTDAGData), nil
	}

	blockGHOSTDAGDataBytes, err := dbContext.Get(gds.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	blockGHOSTDAGData, err := gds.deserializeBlockGHOSTDAGData(blockGHOSTDAGDataBytes)
	if err != nil {
		return nil, err
	}
	gds.cache.Add(blockHash, blockGHOSTDAGData)
	return blockGHOSTDAGData, nil
}

func (gds *ghostdagDataStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return gds.bucket.Key(hash.ByteSlice())
}

func (gds *ghostdagDataStore) serializeBlockGHOSTDAGData(blockGHOSTDAGData *model.BlockGHOSTDAGData) ([]byte, error) {
	return proto.Marshal(serialization.BlockGHOSTDAGDataToDBBlockGHOSTDAGData(blockGHOSTDAGData))
}

func (gds *ghostdagDataStore) deserializeBlockGHOSTDAGData(blockGHOSTDAGDataBytes []byte) (*model.BlockGHOSTDAGData, error) {
	dbBlockGHOSTDAGData := &serialization.DbBlockGhostdagData{}
	err := proto.Unmarshal(blockGHOSTDAGDataBytes, dbBlockGHOSTDAGData)
	if err != nil {
		return nil, err
	}

	return serialization.DBBlockGHOSTDAGDataToBlockGHOSTDAGData(dbBlockGHOSTDAGData)
}
