package blockstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/pkg/errors"
)

var bucket = database.MakeBucket([]byte("blocks"))
var countKey = database.MakeBucket(nil).Key([]byte("blocks-count"))

// blockStore represents a store of blocks
type blockStore struct {
	cache       *lrucache.LRUCache
	countCached uint64
}

// New instantiates a new BlockStore
func New(dbContext model.DBReader, cacheSize int, preallocate bool) (model.BlockStore, error) {
	blockStore := &blockStore{
		cache: lrucache.New(cacheSize, preallocate),
	}

	err := blockStore.initializeCount(dbContext)
	if err != nil {
		return nil, err
	}

	return blockStore, nil
}

func (bs *blockStore) initializeCount(dbContext model.DBReader) error {
	count := uint64(0)
	hasCountBytes, err := dbContext.Has(countKey)
	if err != nil {
		return err
	}
	if hasCountBytes {
		countBytes, err := dbContext.Get(countKey)
		if err != nil {
			return err
		}
		count, err = bs.deserializeBlockCount(countBytes)
		if err != nil {
			return err
		}
	}
	bs.countCached = count
	return nil
}

// Stage stages the given block for the given blockHash
func (bs *blockStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, block *externalapi.DomainBlock) {
	stagingShard := bs.stagingShard(stagingArea)
	stagingShard.toAdd[*blockHash] = block.Clone()
}

func (bs *blockStore) IsStaged(stagingArea *model.StagingArea) bool {
	return bs.stagingShard(stagingArea).isStaged()
}

// Block gets the block associated with the given blockHash
func (bs *blockStore) Block(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	stagingShard := bs.stagingShard(stagingArea)

	return bs.block(dbContext, stagingShard, blockHash)
}

func (bs *blockStore) block(dbContext model.DBReader, stagingShard *blockStagingShard, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	if block, ok := stagingShard.toAdd[*blockHash]; ok {
		return block.Clone(), nil
	}

	if block, ok := bs.cache.Get(blockHash); ok {
		return block.(*externalapi.DomainBlock).Clone(), nil
	}

	blockBytes, err := dbContext.Get(bs.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	block, err := bs.deserializeBlock(blockBytes)
	if err != nil {
		return nil, err
	}
	bs.cache.Add(blockHash, block)
	return block.Clone(), nil
}

// HasBlock returns whether a block with a given hash exists in the store.
func (bs *blockStore) HasBlock(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	stagingShard := bs.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		return true, nil
	}

	if bs.cache.Has(blockHash) {
		return true, nil
	}

	exists, err := dbContext.Has(bs.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Blocks gets the blocks associated with the given blockHashes
func (bs *blockStore) Blocks(dbContext model.DBReader, stagingArea *model.StagingArea, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error) {
	stagingShard := bs.stagingShard(stagingArea)

	blocks := make([]*externalapi.DomainBlock, len(blockHashes))
	for i, hash := range blockHashes {
		var err error
		blocks[i], err = bs.block(dbContext, stagingShard, hash)
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

// Delete deletes the block associated with the given blockHash
func (bs *blockStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	stagingShard := bs.stagingShard(stagingArea)

	if _, ok := stagingShard.toAdd[*blockHash]; ok {
		delete(stagingShard.toAdd, *blockHash)
		return
	}
	stagingShard.toDelete[*blockHash] = struct{}{}
}

func (bs *blockStore) serializeBlock(block *externalapi.DomainBlock) ([]byte, error) {
	dbBlock := serialization.DomainBlockToDbBlock(block)
	return proto.Marshal(dbBlock)
}

func (bs *blockStore) deserializeBlock(blockBytes []byte) (*externalapi.DomainBlock, error) {
	dbBlock := &serialization.DbBlock{}
	err := proto.Unmarshal(blockBytes, dbBlock)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockToDomainBlock(dbBlock)
}

func (bs *blockStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash.ByteSlice())
}

func (bs *blockStore) Count(stagingArea *model.StagingArea) uint64 {
	stagingShard := bs.stagingShard(stagingArea)
	return bs.count(stagingShard)
}

func (bs *blockStore) count(stagingShard *blockStagingShard) uint64 {
	return bs.countCached + uint64(len(stagingShard.toAdd)) - uint64(len(stagingShard.toDelete))
}

func (bs *blockStore) deserializeBlockCount(countBytes []byte) (uint64, error) {
	dbBlockCount := &serialization.DbBlockCount{}
	err := proto.Unmarshal(countBytes, dbBlockCount)
	if err != nil {
		return 0, err
	}
	return dbBlockCount.Count, nil
}

func (bs *blockStore) serializeBlockCount(count uint64) ([]byte, error) {
	dbBlockCount := &serialization.DbBlockCount{Count: count}
	return proto.Marshal(dbBlockCount)
}

type allBlockHashesIterator struct {
	cursor   model.DBCursor
	isClosed bool
}

func (a allBlockHashesIterator) First() bool {
	if a.isClosed {
		panic("Tried using a closed AllBlockHashesIterator")
	}
	return a.cursor.First()
}

func (a allBlockHashesIterator) Next() bool {
	if a.isClosed {
		panic("Tried using a closed AllBlockHashesIterator")
	}
	return a.cursor.Next()
}

func (a allBlockHashesIterator) Get() (*externalapi.DomainHash, error) {
	if a.isClosed {
		return nil, errors.New("Tried using a closed AllBlockHashesIterator")
	}
	key, err := a.cursor.Key()
	if err != nil {
		return nil, err
	}

	blockHashBytes := key.Suffix()
	return externalapi.NewDomainHashFromByteSlice(blockHashBytes)
}

func (a allBlockHashesIterator) Close() error {
	if a.isClosed {
		return errors.New("Tried using a closed AllBlockHashesIterator")
	}
	a.isClosed = true
	err := a.cursor.Close()
	if err != nil {
		return err
	}
	a.cursor = nil
	return nil
}

func (bs *blockStore) AllBlockHashesIterator(dbContext model.DBReader) (model.BlockIterator, error) {
	cursor, err := dbContext.Cursor(bucket)
	if err != nil {
		return nil, err
	}

	return &allBlockHashesIterator{cursor: cursor}, nil
}
