package headersselectedchainstore

import (
	"encoding/binary"

	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucacheuint64tohash"
	"github.com/pkg/errors"
)

var bucketChainBlockHashByIndex = database.MakeBucket([]byte("chain-block-hash-by-index"))
var bucketChainBlockIndexByHash = database.MakeBucket([]byte("chain-block-index-by-hash"))
var highestChainBlockIndexKey = database.MakeBucket(nil).Key([]byte("highest-chain-block-index"))

type headersSelectedChainStore struct {
	cacheByIndex                *lrucacheuint64tohash.LRUCache
	cacheByHash                 *lrucache.LRUCache
	cacheHighestChainBlockIndex uint64
}

// New instantiates a new HeadersSelectedChainStore
func New(cacheSize int, preallocate bool) model.HeadersSelectedChainStore {
	return &headersSelectedChainStore{
		cacheByIndex: lrucacheuint64tohash.New(cacheSize, preallocate),
		cacheByHash:  lrucache.New(cacheSize, preallocate),
	}
}

// Stage stages the given chain changes
func (hscs *headersSelectedChainStore) Stage(dbContext model.DBReader, stagingArea *model.StagingArea, chainChanges *externalapi.SelectedChainPath) error {
	stagingShard := hscs.stagingShard(stagingArea)

	if hscs.IsStaged(stagingArea) {
		return errors.Errorf("can't stage when there's already staged data")
	}

	for _, blockHash := range chainChanges.Removed {
		index, err := hscs.GetIndexByHash(dbContext, stagingArea, blockHash)
		if err != nil {
			return err
		}

		stagingShard.removedByIndex[index] = struct{}{}
		stagingShard.removedByHash[*blockHash] = struct{}{}
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
		stagingShard.addedByIndex[currentIndex] = blockHash
		stagingShard.addedByHash[*blockHash] = currentIndex
		currentIndex++
	}

	return nil
}

func (hscs *headersSelectedChainStore) IsStaged(stagingArea *model.StagingArea) bool {
	return hscs.stagingShard(stagingArea).isStaged()
}

// Get gets the chain block index for the given blockHash
func (hscs *headersSelectedChainStore) GetIndexByHash(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	stagingShard := hscs.stagingShard(stagingArea)

	if index, ok := stagingShard.addedByHash[*blockHash]; ok {
		return index, nil
	}

	if _, ok := stagingShard.removedByHash[*blockHash]; ok {
		return 0, errors.Wrapf(database.ErrNotFound, "couldn't find block %s", blockHash)
	}

	if index, ok := hscs.cacheByHash.Get(blockHash); ok {
		return index.(uint64), nil
	}

	indexBytes, err := dbContext.Get(hscs.hashAsKey(blockHash))
	if err != nil {
		return 0, err
	}

	index, err := hscs.deserializeIndex(indexBytes)
	if err != nil {
		return 0, err
	}

	hscs.cacheByHash.Add(blockHash, index)
	return index, nil
}

func (hscs *headersSelectedChainStore) GetHashByIndex(dbContext model.DBReader, stagingArea *model.StagingArea, index uint64) (*externalapi.DomainHash, error) {
	stagingShard := hscs.stagingShard(stagingArea)

	if blockHash, ok := stagingShard.addedByIndex[index]; ok {
		return blockHash, nil
	}

	if _, ok := stagingShard.removedByIndex[index]; ok {
		return nil, errors.Wrapf(database.ErrNotFound, "couldn't find chain block with index %d", index)
	}

	if blockHash, ok := hscs.cacheByIndex.Get(index); ok {
		return blockHash, nil
	}

	hashBytes, err := dbContext.Get(hscs.indexAsKey(index))
	if err != nil {
		return nil, err
	}

	blockHash, err := binaryserialization.DeserializeHash(hashBytes)
	if err != nil {
		return nil, err
	}
	hscs.cacheByIndex.Add(index, blockHash)
	return blockHash, nil
}

func (hscs *headersSelectedChainStore) serializeIndex(index uint64) []byte {
	return binaryserialization.SerializeUint64(index)
}

func (hscs *headersSelectedChainStore) deserializeIndex(indexBytes []byte) (uint64, error) {
	return binaryserialization.DeserializeUint64(indexBytes)
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

	indexBytes, err := dbContext.Get(highestChainBlockIndexKey)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}

	index, err := hscs.deserializeIndex(indexBytes)
	if err != nil {
		return 0, false, err
	}

	hscs.cacheHighestChainBlockIndex = index
	return index, true, nil
}
