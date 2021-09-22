package finalitystore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/util/staging"
)

var bucketName = []byte("finality-points")

type finalityStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New instantiates a new FinalityStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.FinalityStore {
	return &finalityStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

func (fs *finalityStore) StageFinalityPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, finalityPointHash *externalapi.DomainHash) {
	stagingShard := fs.stagingShard(stagingArea)

	stagingShard.toAdd[*blockHash] = finalityPointHash
}

func (fs *finalityStore) FinalityPoint(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	stagingShard := fs.stagingShard(stagingArea)

	if finalityPointHash, ok := stagingShard.toAdd[*blockHash]; ok {
		return finalityPointHash, nil
	}

	if finalityPointHash, ok := fs.cache.Get(blockHash); ok {
		return finalityPointHash.(*externalapi.DomainHash), nil
	}

	finalityPointHashBytes, err := dbContext.Get(fs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}
	finalityPointHash, err := externalapi.NewDomainHashFromByteSlice(finalityPointHashBytes)
	if err != nil {
		return nil, err
	}

	fs.cache.Add(blockHash, finalityPointHash)
	return finalityPointHash, nil
}

func (fs *finalityStore) IsStaged(stagingArea *model.StagingArea) bool {
	return fs.stagingShard(stagingArea).isStaged()
}

func (fs *finalityStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return fs.bucket.Key(hash.ByteSlice())
}
