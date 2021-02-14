package blockstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
)

var bucket = database.MakeBucket([]byte("blocks"))
var countKey = database.MakeBucket(nil).Key([]byte("blocks-count"))

// blockStore represents a store of blocks
type blockStore struct {
	staging  map[externalapi.DomainHash]*externalapi.DomainBlock
	toDelete map[externalapi.DomainHash]struct{}
	cache    *lrucache.LRUCache
	count    uint64
}

// New instantiates a new BlockStore
func New(dbContext model.DBReader, cacheSize int, preallocate bool) (model.BlockStore, error) {
	blockStore := &blockStore{
		staging:  make(map[externalapi.DomainHash]*externalapi.DomainBlock),
		toDelete: make(map[externalapi.DomainHash]struct{}),
		cache:    lrucache.New(cacheSize, preallocate),
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
	bs.count = count
	return nil
}

// Stage stages the given block for the given blockHash
func (bs *blockStore) Stage(blockHash *externalapi.DomainHash, block *externalapi.DomainBlock) {
	bs.staging[*blockHash] = block.Clone()
}

func (bs *blockStore) IsStaged() bool {
	return len(bs.staging) != 0 || len(bs.toDelete) != 0
}

func (bs *blockStore) Discard() {
	bs.staging = make(map[externalapi.DomainHash]*externalapi.DomainBlock)
	bs.toDelete = make(map[externalapi.DomainHash]struct{})
}

func (bs *blockStore) Commit(dbTx model.DBTransaction) error {
	for hash, block := range bs.staging {
		blockBytes, err := bs.serializeBlock(block)
		if err != nil {
			return err
		}
		err = dbTx.Put(bs.hashAsKey(&hash), blockBytes)
		if err != nil {
			return err
		}
		bs.cache.Add(&hash, block)
	}

	for hash := range bs.toDelete {
		err := dbTx.Delete(bs.hashAsKey(&hash))
		if err != nil {
			return err
		}
		bs.cache.Remove(&hash)
	}

	err := bs.commitCount(dbTx)
	if err != nil {
		return err
	}

	bs.Discard()
	return nil
}

// Block gets the block associated with the given blockHash
func (bs *blockStore) Block(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	if block, ok := bs.staging[*blockHash]; ok {
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
func (bs *blockStore) HasBlock(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := bs.staging[*blockHash]; ok {
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
func (bs *blockStore) Blocks(dbContext model.DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlock, error) {
	blocks := make([]*externalapi.DomainBlock, len(blockHashes))
	for i, hash := range blockHashes {
		var err error
		blocks[i], err = bs.Block(dbContext, hash)
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

// Delete deletes the block associated with the given blockHash
func (bs *blockStore) Delete(blockHash *externalapi.DomainHash) {
	if _, ok := bs.staging[*blockHash]; ok {
		delete(bs.staging, *blockHash)
		return
	}
	bs.toDelete[*blockHash] = struct{}{}
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

func (bs *blockStore) Count() uint64 {
	return bs.count + uint64(len(bs.staging)) - uint64(len(bs.toDelete))
}

func (bs *blockStore) deserializeBlockCount(countBytes []byte) (uint64, error) {
	dbBlockCount := &serialization.DbBlockCount{}
	err := proto.Unmarshal(countBytes, dbBlockCount)
	if err != nil {
		return 0, err
	}
	return dbBlockCount.Count, nil
}

func (bs *blockStore) commitCount(dbTx model.DBTransaction) error {
	count := bs.Count()
	countBytes, err := bs.serializeBlockCount(count)
	if err != nil {
		return err
	}
	err = dbTx.Put(countKey, countBytes)
	if err != nil {
		return err
	}
	bs.count = count
	return nil
}

func (bs *blockStore) serializeBlockCount(count uint64) ([]byte, error) {
	dbBlockCount := &serialization.DbBlockCount{Count: count}
	return proto.Marshal(dbBlockCount)
}

type allBlockHashesIterator struct {
	cursor model.DBCursor
}

func (a allBlockHashesIterator) First() bool {
	return a.cursor.First()
}

func (a allBlockHashesIterator) Next() bool {
	return a.cursor.Next()
}

func (a allBlockHashesIterator) Get() (*externalapi.DomainHash, error) {
	key, err := a.cursor.Key()
	if err != nil {
		return nil, err
	}

	blockHashBytes := key.Suffix()
	return externalapi.NewDomainHashFromByteSlice(blockHashBytes)
}

func (bs *blockStore) AllBlockHashesIterator(dbContext model.DBReader) (model.BlockIterator, error) {
	cursor, err := dbContext.Cursor(bucket)
	if err != nil {
		return nil, err
	}

	return &allBlockHashesIterator{cursor: cursor}, nil
}
