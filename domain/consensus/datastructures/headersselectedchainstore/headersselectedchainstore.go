package headersselectedchainstore

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucacheuint64tohash"
	"github.com/pkg/errors"
)

var bucketChainBlockHashByIndex = dbkeys.MakeBucket([]byte("chain-block-hash-by-index"))
var bucketChainBlockIndexByHash = dbkeys.MakeBucket([]byte("chain-block-index-by-hash"))
var highestChainBlockIndexKey = dbkeys.MakeBucket().Key([]byte("highest-chain-block-index"))

type headersSelectedChainStore struct {
	stagingAddedByHash          map[externalapi.DomainHash]uint64
	stagingRemovedByHash        map[externalapi.DomainHash]struct{}
	stagingAddedByIndex         map[uint64]*externalapi.DomainHash
	stagingRemovedByIndex       map[uint64]struct{}
	cacheByIndex                *lrucacheuint64tohash.LRUCache
	cacheByHash                 *lrucache.LRUCache
	cacheHighestChainBlockIndex uint64
}

// New instantiates a new HeadersSelectedChainStore
func New(cacheSize int) model.HeadersSelectedChainStore {
	return &headersSelectedChainStore{
		stagingAddedByHash:    make(map[externalapi.DomainHash]uint64),
		stagingRemovedByHash:  make(map[externalapi.DomainHash]struct{}),
		stagingAddedByIndex:   make(map[uint64]*externalapi.DomainHash),
		stagingRemovedByIndex: make(map[uint64]struct{}),
		cacheByIndex:          lrucacheuint64tohash.New(cacheSize),
		cacheByHash:           lrucache.New(cacheSize),
	}
}

// Stage stages the given chain changes
func (hscs *headersSelectedChainStore) Stage(dbContext model.DBReader,
	chainChanges *externalapi.SelectedParentChainChanges) error {

	if hscs.IsStaged() {
		return errors.Errorf("can't stage when there's already staged data")
	}

	for _, blockHash := range chainChanges.Removed {
		index, exists, err := hscs.GetIndexByHash(dbContext, blockHash)
		if err != nil {
			return err
		}

		if !exists {
			return errors.Errorf("couldn't find index of %s", blockHash)
		}

		hscs.stagingRemovedByIndex[index] = struct{}{}
		hscs.stagingRemovedByHash[*blockHash] = struct{}{}
	}

	currentIndex := uint64(0)
	highestChainBlockIndex, exists, err := hscs.highestChainBlockIndex(dbContext)
	if err != nil {
		return err
	}

	if exists {
		currentIndex = highestChainBlockIndex - uint64(len(chainChanges.Removed)) + 1
	}

	for _, blockHash := range chainChanges.Added {
		hscs.stagingAddedByIndex[currentIndex] = blockHash
		hscs.stagingAddedByHash[*blockHash] = currentIndex
		currentIndex++
	}

	return nil
}

func (hscs *headersSelectedChainStore) IsStaged() bool {
	return len(hscs.stagingAddedByHash) != 0 ||
		len(hscs.stagingRemovedByHash) != 0 ||
		len(hscs.stagingAddedByIndex) != 0 ||
		len(hscs.stagingAddedByIndex) != 0
}

func (hscs *headersSelectedChainStore) Discard() {
	hscs.stagingAddedByHash = make(map[externalapi.DomainHash]uint64)
	hscs.stagingRemovedByHash = make(map[externalapi.DomainHash]struct{})
	hscs.stagingAddedByIndex = make(map[uint64]*externalapi.DomainHash)
	hscs.stagingRemovedByIndex = make(map[uint64]struct{})
}

func (hscs *headersSelectedChainStore) Commit(dbTx model.DBTransaction) error {
	if !hscs.IsStaged() {
		return nil
	}

	for hash := range hscs.stagingRemovedByHash {
		hashCopy := hash
		err := dbTx.Delete(hscs.hashAsKey(&hashCopy))
		if err != nil {
			return err
		}
		hscs.cacheByHash.Remove(&hashCopy)
	}

	for index := range hscs.stagingRemovedByIndex {
		err := dbTx.Delete(hscs.indexAsKey(index))
		if err != nil {
			return err
		}
		hscs.cacheByIndex.Remove(index)
	}

	highestIndex := uint64(0)
	for hash, index := range hscs.stagingAddedByHash {
		hashCopy := hash
		err := dbTx.Put(hscs.hashAsKey(&hashCopy), hscs.serializeIndex(index))
		if err != nil {
			return err
		}

		err = dbTx.Put(hscs.indexAsKey(index), binaryserialization.SerializeHash(&hashCopy))
		if err != nil {
			return err
		}

		hscs.cacheByHash.Add(&hashCopy, index)
		hscs.cacheByIndex.Add(index, &hashCopy)

		if index > highestIndex {
			highestIndex = index
		}
	}

	err := dbTx.Put(highestChainBlockIndexKey, hscs.serializeIndex(highestIndex))
	if err != nil {
		return err
	}

	hscs.cacheHighestChainBlockIndex = highestIndex

	hscs.Discard()
	return nil
}

// Get gets the chain block index for the given blockHash
func (hscs *headersSelectedChainStore) GetIndexByHash(dbContext model.DBReader, blockHash *externalapi.DomainHash) (uint64, bool, error) {
	if index, ok := hscs.stagingAddedByHash[*blockHash]; ok {
		return index, true, nil
	}

	if _, ok := hscs.stagingRemovedByHash[*blockHash]; ok {
		return 0, false, nil
	}

	if index, ok := hscs.cacheByHash.Get(blockHash); ok {
		return index.(uint64), true, nil
	}

	indexBytes, err := dbContext.Get(hscs.hashAsKey(blockHash))
	if err != nil {
		return 0, false, err
	}

	index := hscs.deserializeIndex(indexBytes)
	hscs.cacheByHash.Add(blockHash, index)
	return index, true, nil
}

func (hscs *headersSelectedChainStore) GetHashByIndex(dbContext model.DBReader, index uint64) (*externalapi.DomainHash, bool, error) {
	if blockHash, ok := hscs.stagingAddedByIndex[index]; ok {
		return blockHash, true, nil
	}

	if _, ok := hscs.stagingRemovedByIndex[index]; ok {
		return nil, false, nil
	}

	if blockHash, ok := hscs.cacheByIndex.Get(index); ok {
		return blockHash, true, nil
	}

	has, err := dbContext.Has(hscs.indexAsKey(index))
	if err != nil {
		return nil, false, err
	}

	if !has {
		return nil, false, nil
	}

	hashBytes, err := dbContext.Get(hscs.indexAsKey(index))
	if err != nil {
		return nil, false, err
	}

	blockHash, err := binaryserialization.DeserializeHash(hashBytes)
	if err != nil {
		return nil, false, err
	}
	hscs.cacheByIndex.Add(index, blockHash)
	return blockHash, true, nil
}

func (hscs *headersSelectedChainStore) serializeIndex(index uint64) []byte {
	return binaryserialization.SerializeChainBlockIndex(index)
}

func (hscs *headersSelectedChainStore) deserializeIndex(indexBytes []byte) uint64 {
	return binaryserialization.DeserializeChainBlockIndex(indexBytes)
}

func (hscs *headersSelectedChainStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return bucketChainBlockIndexByHash.Key(hash.ByteSlice())
}

func (hscs *headersSelectedChainStore) indexAsKey(index uint64) model.DBKey {
	var keyBytes [8]byte
	binary.BigEndian.PutUint64(keyBytes[:], index)
	return bucketChainBlockHashByIndex.Key(keyBytes[:])
}

func (hscs *headersSelectedChainStore) highestChainBlockIndex(dbContext model.DBReader) (uint64, bool, error) {
	if hscs.cacheHighestChainBlockIndex != 0 {
		return hscs.cacheHighestChainBlockIndex, true, nil
	}

	has, err := dbContext.Has(highestChainBlockIndexKey)
	if err != nil {
		return 0, false, err
	}

	if !has {
		return 0, false, nil
	}

	indexBytes, err := dbContext.Get(highestChainBlockIndexKey)
	if err != nil {
		return 0, false, err
	}

	index := hscs.deserializeIndex(indexBytes)
	hscs.cacheHighestChainBlockIndex = index
	return index, true, nil
}
