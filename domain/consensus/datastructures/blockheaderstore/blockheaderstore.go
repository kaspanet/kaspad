package blockheaderstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("block-headers"))
var countKey = dbkeys.MakeBucket().Key([]byte("block-headers-count"))

// blockHeaderStore represents a store of blocks
type blockHeaderStore struct {
	staging  map[externalapi.DomainHash]*externalapi.DomainBlockHeader
	toDelete map[externalapi.DomainHash]struct{}
	count    uint64
}

// New instantiates a new BlockHeaderStore
func New(dbContext model.DBReader) (model.BlockHeaderStore, error) {
	blockHeaderStore := &blockHeaderStore{
		staging:  make(map[externalapi.DomainHash]*externalapi.DomainBlockHeader),
		toDelete: make(map[externalapi.DomainHash]struct{}),
	}

	err := blockHeaderStore.initializeCount(dbContext)
	if err != nil {
		return nil, err
	}

	return blockHeaderStore, nil
}

func (bms *blockHeaderStore) initializeCount(dbContext model.DBReader) error {
	var count uint64
	hasCountBytes, err := dbContext.Has(countKey)
	if err != nil {
		return err
	}
	if hasCountBytes {
		countBytes, err := dbContext.Get(countKey)
		if err != nil {
			return err
		}
		count, err = bms.deserializeHeaderCount(countBytes)
		if err != nil {
			return err
		}
	}
	bms.count = count
	return nil
}

// Stage stages the given block header for the given blockHash
func (bms *blockHeaderStore) Stage(blockHash *externalapi.DomainHash, blockHeader *externalapi.DomainBlockHeader) error {
	clone, err := bms.cloneHeader(blockHeader)
	if err != nil {
		return err
	}

	bms.staging[*blockHash] = clone
	return nil
}

func (bms *blockHeaderStore) IsStaged() bool {
	return len(bms.staging) != 0 || len(bms.toDelete) != 0
}

func (bms *blockHeaderStore) Discard() {
	bms.staging = make(map[externalapi.DomainHash]*externalapi.DomainBlockHeader)
	bms.toDelete = make(map[externalapi.DomainHash]struct{})
}

func (bms *blockHeaderStore) Commit(dbTx model.DBTransaction) error {
	for hash, header := range bms.staging {
		headerBytes, err := bms.serializeHeader(header)
		if err != nil {
			return err
		}
		err = dbTx.Put(bms.hashAsKey(&hash), headerBytes)
		if err != nil {
			return err
		}
	}

	for hash := range bms.toDelete {
		err := dbTx.Delete(bms.hashAsKey(&hash))
		if err != nil {
			return err
		}
	}

	err := bms.commitCount(dbTx)
	if err != nil {
		return err
	}

	bms.Discard()
	return nil
}

// BlockHeader gets the block header associated with the given blockHash
func (bms *blockHeaderStore) BlockHeader(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainBlockHeader, error) {
	if header, ok := bms.staging[*blockHash]; ok {
		return header, nil
	}

	headerBytes, err := dbContext.Get(bms.hashAsKey(blockHash))
	if err != nil {
		return nil, err
	}

	return bms.deserializeHeader(headerBytes)
}

// HasBlock returns whether a block header with a given hash exists in the store.
func (bms *blockHeaderStore) HasBlockHeader(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := bms.staging[*blockHash]; ok {
		return true, nil
	}

	exists, err := dbContext.Has(bms.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

// BlockHeaders gets the block headers associated with the given blockHashes
func (bms *blockHeaderStore) BlockHeaders(dbContext model.DBReader, blockHashes []*externalapi.DomainHash) ([]*externalapi.DomainBlockHeader, error) {
	headers := make([]*externalapi.DomainBlockHeader, len(blockHashes))
	for i, hash := range blockHashes {
		var err error
		headers[i], err = bms.BlockHeader(dbContext, hash)
		if err != nil {
			return nil, err
		}
	}
	return headers, nil
}

// Delete deletes the block associated with the given blockHash
func (bms *blockHeaderStore) Delete(blockHash *externalapi.DomainHash) {
	if _, ok := bms.staging[*blockHash]; ok {
		delete(bms.staging, *blockHash)
		return
	}
	bms.toDelete[*blockHash] = struct{}{}
}

func (bms *blockHeaderStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}

func (bms *blockHeaderStore) serializeHeader(header *externalapi.DomainBlockHeader) ([]byte, error) {
	dbBlockHeader := serialization.DomainBlockHeaderToDbBlockHeader(header)
	return proto.Marshal(dbBlockHeader)
}

func (bms *blockHeaderStore) deserializeHeader(headerBytes []byte) (*externalapi.DomainBlockHeader, error) {
	dbBlockHeader := &serialization.DbBlockHeader{}
	err := proto.Unmarshal(headerBytes, dbBlockHeader)
	if err != nil {
		return nil, err
	}
	return serialization.DbBlockHeaderToDomainBlockHeader(dbBlockHeader)
}

func (bms *blockHeaderStore) cloneHeader(header *externalapi.DomainBlockHeader) (*externalapi.DomainBlockHeader, error) {
	serialized, err := bms.serializeHeader(header)
	if err != nil {
		return nil, err
	}

	return bms.deserializeHeader(serialized)
}

func (bms *blockHeaderStore) Count() uint64 {
	return bms.count + uint64(len(bms.staging)) - uint64(len(bms.toDelete))
}

func (bms *blockHeaderStore) deserializeHeaderCount(countBytes []byte) (uint64, error) {
	dbBlockHeaderCount := &serialization.DbBlockHeaderCount{}
	err := proto.Unmarshal(countBytes, dbBlockHeaderCount)
	if err != nil {
		return 0, err
	}
	return dbBlockHeaderCount.Count, nil
}

func (bms *blockHeaderStore) commitCount(dbTx model.DBTransaction) error {
	count := bms.Count()
	countBytes, err := bms.serializeHeaderCount(count)
	if err != nil {
		return err
	}
	return dbTx.Put(countKey, countBytes)
}

func (bms *blockHeaderStore) serializeHeaderCount(count uint64) ([]byte, error) {
	dbBlockHeaderCount := &serialization.DbBlockHeaderCount{Count: count}
	return proto.Marshal(dbBlockHeaderCount)
}
