package blockrelationstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/util/staging"
)

var bucketName = []byte("block-relations")

// blockRelationStore represents a store of BlockRelations
type blockRelationStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New instantiates a new BlockRelationStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.BlockRelationStore {
	return &blockRelationStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

func (brs *blockRelationStore) StageBlockRelation(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, blockRelations *model.BlockRelations) {
	stagingShard := brs.stagingShard(stagingArea)

	stagingShard.toAdd[*blockHash] = blockRelations.Clone()
}

func (brs *blockRelationStore) IsStaged(stagingArea *model.StagingArea) bool {
	return brs.stagingShard(stagingArea).isStaged()
}

func (brs *blockRelationStore) BlockRelation(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*model.BlockRelations, error) {
	stagingShard := brs.stagingShard(stagingArea)

	if blockRelations, ok := stagingShard.toAdd[*blockHash]; ok {
		return blockRelations.Clone(), nil
	}

	if blockRelations, ok := brs.cache.Get(blockHash); ok {
		return blockRelations.(*model.BlockRelations).Clone(), nil
	}

	blockRelationsBytes, err := dbContext.Get(brs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	blockRelations, err := brs.deserializeBlockRelations(blockRelationsBytes)
	if err != nil {
		return nil, err
	}
	brs.cache.Add(blockHash, blockRelations)
	return blockRelations.Clone(), nil
}

func (brs *blockRelationStore) Has(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	stagingShard := brs.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		return true, nil
	}

	if brs.cache.Has(blockHash) {
		return true, nil
	}

	return dbContext.Has(brs.hashAsKey(blockHash))
}

func (brs *blockRelationStore) UnstageAll(stagingArea *model.StagingArea) {
	stagingShard := brs.stagingShard(stagingArea)
	stagingShard.toAdd = make(map[externalapi.DomainHash]*model.BlockRelations)
}

func (brs *blockRelationStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return brs.bucket.Key(hash.ByteSlice())
}

func (brs *blockRelationStore) serializeBlockRelations(blockRelations *model.BlockRelations) ([]byte, error) {
	dbBlockRelations := serialization.DomainBlockRelationsToDbBlockRelations(blockRelations)
	return proto.Marshal(dbBlockRelations)
}

func (brs *blockRelationStore) deserializeBlockRelations(blockRelationsBytes []byte) (*model.BlockRelations, error) {
	dbBlockRelations := &serialization.DbBlockRelations{}
	err := proto.Unmarshal(blockRelationsBytes, dbBlockRelations)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockRelationsToDomainBlockRelations(dbBlockRelations)
}
