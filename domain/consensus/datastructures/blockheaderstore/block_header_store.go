package blockheaderstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/domain/prefixmanager"
)

var bucketName = []byte("block-headers")
var countKeyName = []byte("block-headers-count")

// blockHeaderStore represents a store of blocks
type blockHeaderStore struct {
	cache       *lrucache.LRUCache
	countCached uint64
	bucket      model.DBBucket
	countKey    model.DBKey
}

// New instantiates a new BlockHeaderStore
func New(dbContext model.DBReader, prefix *prefixmanager.Prefix, cacheSize int, preallocate bool) (model.BlockHeaderStore, error) {
	blockHeaderStore := &blockHeaderStore{
		cache:    lrucache.New(cacheSize, preallocate),
		bucket:   database.MakeBucket(prefix.Serialize()).Bucket(bucketName),
		countKey: database.MakeBucket(prefix.Serialize()).Key(countKeyName),
	}

	err := blockHeaderStore.initializeCount(dbContext)
	if err != nil {
		return nil, err
	}

	return blockHeaderStore, nil
}

func (bhs *blockHeaderStore) initializeCount(dbContext model.DBReader) error {
	count := uint64(0)
	hasCountBytes, err := dbContext.Has(bhs.countKey)
	if err != nil {
		return err
	}
	if hasCountBytes {
		countBytes, err := dbContext.Get(bhs.countKey)
		if err != nil {
			return err
		}
		count, err = bhs.deserializeHeaderCount(countBytes)
		if err != nil {
			return err
		}
	}
	bhs.countCached = count
	return nil
}

// Stage stages the given block header for the given blockHash
func (bhs *blockHeaderStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, blockHeader externalapi.BlockHeader) {
	stagingShard := bhs.stagingShard(stagingArea)
	stagingShard.toAdd[*blockHash] = blockHeader
}

func (bhs *blockHeaderStore) IsStaged(stagingArea *model.StagingArea) bool {
	return bhs.stagingShard(stagingArea).isStaged()
}

// BlockHeader gets the block header associated with the given blockHash
func (bhs *blockHeaderStore) BlockHeader(dbContext model.DBReader, stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {

	stagingShard := bhs.stagingShard(stagingArea)

	return bhs.blockHeader(dbContext, stagingShard, blockHash)
}

func (bhs *blockHeaderStore) blockHeader(dbContext model.DBReader, stagingShard *blockHeaderStagingShard,
	blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {

	if header, ok := stagingShard.toAdd[*blockHash]; ok {
		return header, nil
	}

	if header, ok := bhs.cache.Get(blockHash); ok {
		return header.(externalapi.BlockHeader), nil
	}

	headerBytes, err := dbContext.Get(bhs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	header, err := bhs.deserializeHeader(headerBytes)
	if err != nil {
		return nil, err
	}
	bhs.cache.Add(blockHash, header)
	return header, nil
}

// HasBlock returns whether a block header with a given hash exists in the store.
func (bhs *blockHeaderStore) HasBlockHeader(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	stagingShard := bhs.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		return true, nil
	}

	if bhs.cache.Has(blockHash) {
		return true, nil
	}

	exists, err := dbContext.Has(bhs.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

// BlockHeaders gets the block headers associated with the given blockHashes
func (bhs *blockHeaderStore) BlockHeaders(dbContext model.DBReader, stagingArea *model.StagingArea,
	blockHashes []*externalapi.DomainHash) ([]externalapi.BlockHeader, error) {

	stagingShard := bhs.stagingShard(stagingArea)

	headers := make([]externalapi.BlockHeader, len(blockHashes))
	for i, hash := range blockHashes {
		var err error
		headers[i], err = bhs.blockHeader(dbContext, stagingShard, hash)
		if err != nil {
			return nil, err
		}
	}
	return headers, nil
}

// Delete deletes the block associated with the given blockHash
func (bhs *blockHeaderStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	stagingShard := bhs.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		delete(stagingShard.toAdd, *blockHash)
		return
	}
	stagingShard.toDelete[*blockHash] = struct{}{}
}

func (bhs *blockHeaderStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bhs.bucket.Key(hash.ByteSlice())
}

func (bhs *blockHeaderStore) serializeHeader(header externalapi.BlockHeader) ([]byte, error) {
	dbBlockHeader := serialization.DomainBlockHeaderToDbBlockHeader(header)
	return proto.Marshal(dbBlockHeader)
}

func (bhs *blockHeaderStore) deserializeHeader(headerBytes []byte) (externalapi.BlockHeader, error) {
	dbBlockHeader := &serialization.DbBlockHeader{}
	err := proto.Unmarshal(headerBytes, dbBlockHeader)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockHeaderToDomainBlockHeader(dbBlockHeader)
}

func (bhs *blockHeaderStore) Count(stagingArea *model.StagingArea) uint64 {
	stagingShard := bhs.stagingShard(stagingArea)

	return bhs.count(stagingShard)
}

func (bhs *blockHeaderStore) count(stagingShard *blockHeaderStagingShard) uint64 {
	return bhs.countCached + uint64(len(stagingShard.toAdd)) - uint64(len(stagingShard.toDelete))
}

func (bhs *blockHeaderStore) deserializeHeaderCount(countBytes []byte) (uint64, error) {
	dbBlockHeaderCount := &serialization.DbBlockHeaderCount{}
	err := proto.Unmarshal(countBytes, dbBlockHeaderCount)
	if err != nil {
		return 0, err
	}
	return dbBlockHeaderCount.Count, nil
}

func (bhs *blockHeaderStore) serializeHeaderCount(count uint64) ([]byte, error) {
	dbBlockHeaderCount := &serialization.DbBlockHeaderCount{Count: count}
	return proto.Marshal(dbBlockHeaderCount)
}
