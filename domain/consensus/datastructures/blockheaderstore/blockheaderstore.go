package blockheaderstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
)

var bucket = dbkeys.MakeBucket([]byte("block-headers"))
var countKey = dbkeys.MakeBucket().Key([]byte("block-headers-count"))

// blockHeaderStore represents a store of blocks
type blockHeaderStore struct {
	staging  map[externalapi.DomainHash]*externalapi.DomainBlockHeader
	toDelete map[externalapi.DomainHash]struct{}
	cache    *lrucache.LRUCache
	count    uint64
}

// New instantiates a new BlockHeaderStore
func New(dbContext model.DBReader, cacheSize int) (model.BlockHeaderStore, error) {
	blockHeaderStore := &blockHeaderStore{
		staging:  make(map[externalapi.DomainHash]*externalapi.DomainBlockHeader),
		toDelete: make(map[externalapi.DomainHash]struct{}),
		cache:    lrucache.New(cacheSize),
	}

	err := blockHeaderStore.initializeCount(dbContext)
	if err != nil {
		return nil, err
	}

	return blockHeaderStore, nil
}

func (bhs *blockHeaderStore) initializeCount(dbContext model.DBReader) error {
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
		count, err = bhs.deserializeHeaderCount(countBytes)
		if err != nil {
			return err
		}
	}
	bhs.count = count
	return nil
}

// Stage stages the given block header for the given blockHash
func (bhs *blockHeaderStore) Stage(blockHash *externalapi.DomainHash, blockHeader *externalapi.DomainBlockHeader) {
	bhs.staging[*blockHash] = blockHeader.Clone()
}

func (bhs *blockHeaderStore) IsStaged() bool {
	return len(bhs.staging) != 0 || len(bhs.toDelete) != 0
}

func (bhs *blockHeaderStore) Discard() {
	bhs.staging = make(map[externalapi.DomainHash]*externalapi.DomainBlockHeader)
	bhs.toDelete = make(map[externalapi.DomainHash]struct{})
}

func (bhs *blockHeaderStore) Commit(dbTx model.DBTransaction) error {
	for hash, header := range bhs.staging {
		headerBytes, err := bhs.serializeHeader(header)
		if err != nil {
			return err
		}
		err = dbTx.Put(bhs.hashAsKey(&hash), headerBytes)
		if err != nil {
			return err
		}
		bhs.cache.Add(&hash, header)
	}

	for hash := range bhs.toDelete {
		err := dbTx.Delete(bhs.hashAsKey(&hash))
		if err != nil {
			return err
		}
		bhs.cache.Remove(&hash)
	}

	err := bhs.commitCount(dbTx)
	if err != nil {
		return err
	}

	bhs.Discard()
	return nil
}

// BlockHeader gets the block header associated with the given blockHash
func (bhs *blockHeaderStore) BlockHeader(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	if header, ok := bhs.staging[*blockHash]; ok {
		return header.Clone(), nil
	}

	if header, ok := bhs.cache.Get(blockHash); ok {
		return header.(*externalapi.DomainBlockHeader).Clone(), nil
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
	return header.Clone(), nil
}

// HasBlock returns whether a block header with a given hash exists in the store.
func (bhs *blockHeaderStore) HasBlockHeader(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := bhs.staging[*blockHash]; ok {
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
func (bhs *blockHeaderStore) BlockHeaders(dbContext model.DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlockHeader, error) {
	headers := make([]*externalapi.DomainBlockHeader, len(blockHashes))
	for i, hash := range blockHashes {
		var err error
		headers[i], err = bhs.BlockHeader(dbContext, hash)
		if err != nil {
			return nil, err
		}
	}
	return headers, nil
}

// Delete deletes the block associated with the given blockHash
func (bhs *blockHeaderStore) Delete(blockHash *externalapi.DomainHash) {
	if _, ok := bhs.staging[*blockHash]; ok {
		delete(bhs.staging, *blockHash)
		return
	}
	bhs.toDelete[*blockHash] = struct{}{}
}

func (bhs *blockHeaderStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash.BytesSlice())
}

func (bhs *blockHeaderStore) serializeHeader(header *externalapi.DomainBlockHeader) ([]byte, error) {
	dbBlockHeader := serialization.DomainBlockHeaderToDbBlockHeader(header)
	return proto.Marshal(dbBlockHeader)
}

func (bhs *blockHeaderStore) deserializeHeader(headerBytes []byte) (*externalapi.DomainBlockHeader, error) {
	dbBlockHeader := &serialization.DbBlockHeader{}
	err := proto.Unmarshal(headerBytes, dbBlockHeader)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockHeaderToDomainBlockHeader(dbBlockHeader)
}

func (bhs *blockHeaderStore) Count() uint64 {
	return bhs.count + uint64(len(bhs.staging)) - uint64(len(bhs.toDelete))
}

func (bhs *blockHeaderStore) deserializeHeaderCount(countBytes []byte) (uint64, error) {
	dbBlockHeaderCount := &serialization.DbBlockHeaderCount{}
	err := proto.Unmarshal(countBytes, dbBlockHeaderCount)
	if err != nil {
		return 0, err
	}
	return dbBlockHeaderCount.Count, nil
}

func (bhs *blockHeaderStore) commitCount(dbTx model.DBTransaction) error {
	count := bhs.Count()
	countBytes, err := bhs.serializeHeaderCount(count)
	if err != nil {
		return err
	}
	err = dbTx.Put(countKey, countBytes)
	if err != nil {
		return err
	}
	bhs.count = count
	return nil
}

func (bhs *blockHeaderStore) serializeHeaderCount(count uint64) ([]byte, error) {
	dbBlockHeaderCount := &serialization.DbBlockHeaderCount{Count: count}
	return proto.Marshal(dbBlockHeaderCount)
}
