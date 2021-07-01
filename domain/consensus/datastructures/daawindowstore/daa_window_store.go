package daawindowstore

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucachehashpairtoblockghostdagdatahashpair"
	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"
)

var bucketName = []byte("daa-window")

type daaWindowStore struct {
	cache  *lrucachehashpairtoblockghostdagdatahashpair.LRUCache
	bucket model.DBBucket
}

// New instantiates a new DAAWindowStore
func New(prefix *prefix.Prefix, cacheSize int, preallocate bool) model.DAAWindowStore {
	return &daaWindowStore{
		cache:  lrucachehashpairtoblockghostdagdatahashpair.New(cacheSize, preallocate),
		bucket: database.MakeBucket(prefix.Serialize()).Bucket(bucketName),
	}
}

func (daaws *daaWindowStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, index uint64, pair *externalapi.BlockGHOSTDAGDataHashPair) {
	stagingShard := daaws.stagingShard(stagingArea)

	key := newDBKey(blockHash, index)
	if _, ok := stagingShard.toAdd[key]; !ok {
		stagingShard.toAdd[key] = pair
	}

}

func (daaws *daaWindowStore) DAAWindowBlock(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, index uint64) (*externalapi.BlockGHOSTDAGDataHashPair, error) {
	stagingShard := daaws.stagingShard(stagingArea)

	dbKey := newDBKey(blockHash, index)
	if pair, ok := stagingShard.toAdd[dbKey]; ok {
		return pair, nil
	}

	if pair, ok := daaws.cache.Get(blockHash, index); ok {
		return pair, nil
	}

	pairBytes, err := dbContext.Get(daaws.key(dbKey))
	if err != nil {
		return nil, err
	}

	pair, err := deserializePairBytes(pairBytes)
	if err != nil {
		return nil, err
	}

	daaws.cache.Add(blockHash, index, pair)
	return pair, nil
}

func deserializePairBytes(pairBytes []byte) (*externalapi.BlockGHOSTDAGDataHashPair, error) {
	dbPair := &serialization.DbBlockGHOSTDAGDataHashPair{}
	err := proto.Unmarshal(pairBytes, dbPair)
	if err != nil {
		return nil, err
	}

	return serialization.DbBlockGHOSTDAGDataHashPairToBlockGHOSTDAGDataHashPair(dbPair)
}

func (daaws *daaWindowStore) IsStaged(stagingArea *model.StagingArea) bool {
	return daaws.stagingShard(stagingArea).isStaged()
}

func (daaws *daaWindowStore) key(key dbKey) model.DBKey {
	keyIndexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(keyIndexBytes, key.index)
	return daaws.bucket.Bucket(key.blockHash.ByteSlice()).Key(keyIndexBytes)
}
