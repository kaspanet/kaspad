package daablocksstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
)

var daaScoreBucketName = []byte("daa-score")
var daaAddedBlocksBucketName = []byte("daa-added-blocks")

// daaBlocksStore represents a store of DAABlocksStore
type daaBlocksStore struct {
	daaScoreLRUCache       *lrucache.LRUCache
	daaAddedBlocksLRUCache *lrucache.LRUCache
	daaScoreBucket         model.DBBucket
	daaAddedBlocksBucket   model.DBBucket
}

// New instantiates a new DAABlocksStore
func New(prefix byte, daaScoreCacheSize int, daaAddedBlocksCacheSize int, preallocate bool) model.DAABlocksStore {
	return &daaBlocksStore{
		daaScoreLRUCache:       lrucache.New(daaScoreCacheSize, preallocate),
		daaAddedBlocksLRUCache: lrucache.New(daaAddedBlocksCacheSize, preallocate),
		daaScoreBucket:         database.MakeBucket([]byte{prefix}).Bucket(daaScoreBucketName),
		daaAddedBlocksBucket:   database.MakeBucket([]byte{prefix}).Bucket(daaAddedBlocksBucketName),
	}
}

func (daas *daaBlocksStore) StageDAAScore(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, daaScore uint64) {
	stagingShard := daas.stagingShard(stagingArea)

	stagingShard.daaScoreToAdd[*blockHash] = daaScore
}

func (daas *daaBlocksStore) StageBlockDAAAddedBlocks(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, addedBlocks []*externalapi.DomainHash) {
	stagingShard := daas.stagingShard(stagingArea)

	stagingShard.daaAddedBlocksToAdd[*blockHash] = externalapi.CloneHashes(addedBlocks)
}

func (daas *daaBlocksStore) IsStaged(stagingArea *model.StagingArea) bool {
	return daas.stagingShard(stagingArea).isStaged()
}

func (daas *daaBlocksStore) DAAScore(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	stagingShard := daas.stagingShard(stagingArea)

	if daaScore, ok := stagingShard.daaScoreToAdd[*blockHash]; ok {
		return daaScore, nil
	}

	if daaScore, ok := daas.daaScoreLRUCache.Get(blockHash); ok {
		return daaScore.(uint64), nil
	}

	daaScoreBytes, err := dbContext.Get(daas.daaScoreHashAsKey(blockHash))
	if err != nil {
		return 0, err
	}

	daaScore, err := binaryserialization.DeserializeUint64(daaScoreBytes)
	if err != nil {
		return 0, err
	}
	daas.daaScoreLRUCache.Add(blockHash, daaScore)
	return daaScore, nil
}

func (daas *daaBlocksStore) DAAAddedBlocks(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	stagingShard := daas.stagingShard(stagingArea)

	if addedBlocks, ok := stagingShard.daaAddedBlocksToAdd[*blockHash]; ok {
		return externalapi.CloneHashes(addedBlocks), nil
	}

	if addedBlocks, ok := daas.daaAddedBlocksLRUCache.Get(blockHash); ok {
		return externalapi.CloneHashes(addedBlocks.([]*externalapi.DomainHash)), nil
	}

	addedBlocksBytes, err := dbContext.Get(daas.daaAddedBlocksHashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	addedBlocks, err := binaryserialization.DeserializeHashes(addedBlocksBytes)
	if err != nil {
		return nil, err
	}
	daas.daaAddedBlocksLRUCache.Add(blockHash, addedBlocks)
	return externalapi.CloneHashes(addedBlocks), nil
}

func (daas *daaBlocksStore) daaScoreHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return daas.daaScoreBucket.Key(hash.ByteSlice())
}

func (daas *daaBlocksStore) daaAddedBlocksHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return daas.daaAddedBlocksBucket.Key(hash.ByteSlice())
}

func (daas *daaBlocksStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	stagingShard := daas.stagingShard(stagingArea)

	if _, ok := stagingShard.daaScoreToAdd[*blockHash]; ok {
		delete(stagingShard.daaScoreToAdd, *blockHash)
	} else {
		stagingShard.daaAddedBlocksToDelete[*blockHash] = struct{}{}
	}

	if _, ok := stagingShard.daaAddedBlocksToAdd[*blockHash]; ok {
		delete(stagingShard.daaAddedBlocksToAdd, *blockHash)
	} else {
		stagingShard.daaAddedBlocksToDelete[*blockHash] = struct{}{}
	}
}
