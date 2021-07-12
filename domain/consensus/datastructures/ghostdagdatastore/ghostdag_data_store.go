package ghostdagdatastore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucacheghostdagdata"
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
)

var ghostdagDataBucketName = []byte("block-ghostdag-data")
var metaDataBucketName = []byte("block-with-meta-data-ghostdag-data")

// ghostdagDataStore represents a store of BlockGHOSTDAGData
type ghostdagDataStore struct {
	cache              *lrucacheghostdagdata.LRUCache
	ghostdagDataBucket model.DBBucket
	metaDataBucket     model.DBBucket
}

// New instantiates a new GHOSTDAGDataStore
func New(prefix *prefix.Prefix, cacheSize int, preallocate bool) model.GHOSTDAGDataStore {
	return &ghostdagDataStore{
		cache:              lrucacheghostdagdata.New(cacheSize, preallocate),
		ghostdagDataBucket: database.MakeBucket(prefix.Serialize()).Bucket(ghostdagDataBucketName),
		metaDataBucket:     database.MakeBucket(prefix.Serialize()).Bucket(metaDataBucketName),
	}
}

// Stage stages the given blockGHOSTDAGData for the given blockHash
func (gds *ghostdagDataStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	blockGHOSTDAGData *externalapi.BlockGHOSTDAGData, isMetaData bool) {

	stagingShard := gds.stagingShard(stagingArea)

	stagingShard.toAdd[newKey(blockHash, isMetaData)] = blockGHOSTDAGData
}

func (gds *ghostdagDataStore) IsStaged(stagingArea *model.StagingArea) bool {
	return gds.stagingShard(stagingArea).isStaged()
}

// Get gets the blockGHOSTDAGData associated with the given blockHash
func (gds *ghostdagDataStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isMetaData bool) (*externalapi.BlockGHOSTDAGData, error) {
	stagingShard := gds.stagingShard(stagingArea)

	key := newKey(blockHash, isMetaData)
	if blockGHOSTDAGData, ok := stagingShard.toAdd[key]; ok {
		return blockGHOSTDAGData, nil
	}

	if blockGHOSTDAGData, ok := gds.cache.Get(blockHash, isMetaData); ok {
		return blockGHOSTDAGData, nil
	}

	blockGHOSTDAGDataBytes, err := dbContext.Get(gds.serializeKey(key))
	if err != nil {
		return nil, err
	}

	blockGHOSTDAGData, err := gds.deserializeBlockGHOSTDAGData(blockGHOSTDAGDataBytes)
	if err != nil {
		return nil, err
	}
	gds.cache.Add(blockHash, isMetaData, blockGHOSTDAGData)
	return blockGHOSTDAGData, nil
}

func (gds *ghostdagDataStore) serializeKey(k key) model.DBKey {
	if k.isMetaData {
		return gds.metaDataBucket.Key(k.hash.ByteSlice())
	}
	return gds.ghostdagDataBucket.Key(k.hash.ByteSlice())
}

func (gds *ghostdagDataStore) serializeBlockGHOSTDAGData(blockGHOSTDAGData *externalapi.BlockGHOSTDAGData) ([]byte, error) {
	return proto.Marshal(serialization.BlockGHOSTDAGDataToDBBlockGHOSTDAGData(blockGHOSTDAGData))
}

func (gds *ghostdagDataStore) deserializeBlockGHOSTDAGData(blockGHOSTDAGDataBytes []byte) (*externalapi.BlockGHOSTDAGData, error) {
	dbBlockGHOSTDAGData := &serialization.DbBlockGhostdagData{}
	err := proto.Unmarshal(blockGHOSTDAGDataBytes, dbBlockGHOSTDAGData)
	if err != nil {
		return nil, err
	}

	return serialization.DBBlockGHOSTDAGDataToBlockGHOSTDAGData(dbBlockGHOSTDAGData)
}
