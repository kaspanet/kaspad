package blockstatusstore

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/golang-lru/simplelru"
	"github.com/kaspanet/kaspad/domain/consensus/database/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
)

var bucket = dbkeys.MakeBucket([]byte("block-statuses"))

// blockStatusStore represents a store of BlockStatuses
type blockStatusStore struct {
	staging map[externalapi.DomainHash]externalapi.BlockStatus
	cache   simplelru.LRUCache
}

// New instantiates a new BlockStatusStore
func New(cacheSize int) (model.BlockStatusStore, error) {
	blockStatusStore := &blockStatusStore{
		staging: make(map[externalapi.DomainHash]externalapi.BlockStatus),
	}

	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		return nil, err
	}
	blockStatusStore.cache = cache

	return blockStatusStore, nil
}

// Stage stages the given blockStatus for the given blockHash
func (bss *blockStatusStore) Stage(blockHash *externalapi.DomainHash, blockStatus externalapi.BlockStatus) {
	bss.staging[*blockHash] = blockStatus
}

func (bss *blockStatusStore) IsStaged() bool {
	return len(bss.staging) != 0
}

func (bss *blockStatusStore) Discard() {
	bss.staging = make(map[externalapi.DomainHash]externalapi.BlockStatus)
}

func (bss *blockStatusStore) Commit(dbTx model.DBTransaction) error {
	for hash, status := range bss.staging {
		blockStatusBytes, err := bss.serializeBlockStatus(status)
		if err != nil {
			return err
		}
		err = dbTx.Put(bss.hashAsKey(&hash), blockStatusBytes)
		if err != nil {
			return err
		}
	}

	bss.Discard()
	return nil
}

// Get gets the blockStatus associated with the given blockHash
func (bss *blockStatusStore) Get(dbContext model.DBReader, blockHash *externalapi.DomainHash) (externalapi.BlockStatus, error) {
	if status, ok := bss.staging[*blockHash]; ok {
		return status, nil
	}

	statusBytes, err := dbContext.Get(bss.hashAsKey(blockHash))
	if err != nil {
		return 0, err
	}

	return bss.deserializeBlockStatus(statusBytes)
}

// Exists returns true if the blockStatus for the given blockHash exists
func (bss *blockStatusStore) Exists(dbContext model.DBReader, blockHash *externalapi.DomainHash) (bool, error) {
	if _, ok := bss.staging[*blockHash]; ok {
		return true, nil
	}

	exists, err := dbContext.Has(bss.hashAsKey(blockHash))
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (bss *blockStatusStore) serializeBlockStatus(status externalapi.BlockStatus) ([]byte, error) {
	dbBlockStatus := serialization.DomainBlockStatusToDbBlockStatus(status)
	return proto.Marshal(dbBlockStatus)
}

func (bss *blockStatusStore) deserializeBlockStatus(statusBytes []byte) (externalapi.BlockStatus, error) {
	dbBlockStatus := &serialization.DbBlockStatus{}
	err := proto.Unmarshal(statusBytes, dbBlockStatus)
	if err != nil {
		return 0, err
	}
	return serialization.DbBlockStatusToDomainBlockStatus(dbBlockStatus), nil
}

func (bss *blockStatusStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucket.Key(hash[:])
}
