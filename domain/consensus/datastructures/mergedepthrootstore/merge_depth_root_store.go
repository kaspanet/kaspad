package mergedepthrootstore

import (
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/lrucache"
	"github.com/zoomy-network/zoomyd/util/staging"
)

var bucketName = []byte("merge-depth-roots")

type mergeDepthRootStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New instantiates a new MergeDepthRootStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.MergeDepthRootStore {
	return &mergeDepthRootStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

func (mdrs *mergeDepthRootStore) StageMergeDepthRoot(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, root *externalapi.DomainHash) {
	stagingShard := mdrs.stagingShard(stagingArea)

	stagingShard.toAdd[*blockHash] = root
}

func (mdrs *mergeDepthRootStore) MergeDepthRoot(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	stagingShard := mdrs.stagingShard(stagingArea)

	if root, ok := stagingShard.toAdd[*blockHash]; ok {
		return root, nil
	}

	if root, ok := mdrs.cache.Get(blockHash); ok {
		return root.(*externalapi.DomainHash), nil
	}

	rootBytes, err := dbContext.Get(mdrs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}
	root, err := externalapi.NewDomainHashFromByteSlice(rootBytes)
	if err != nil {
		return nil, err
	}

	mdrs.cache.Add(blockHash, root)
	return root, nil
}

func (mdrs *mergeDepthRootStore) IsStaged(stagingArea *model.StagingArea) bool {
	return mdrs.stagingShard(stagingArea).isStaged()
}

func (mdrs *mergeDepthRootStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return mdrs.bucket.Key(hash.ByteSlice())
}
