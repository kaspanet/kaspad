package daablocksstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/lrucache"
)

var daaScoreBucket = database.MakeBucket([]byte("daa-score"))
var daaAddedBlocksBucket = database.MakeBucket([]byte("daa-added-blocks"))

// daaBlocksStore represents a store of DAABlocksStore
type daaBlocksStore struct {
	daaScoreStaging        map[externalapi.DomainHash]uint64
	daaAddedBlocksStaging  map[externalapi.DomainHash][]*externalapi.DomainHash
	daaScoreToDelete       map[externalapi.DomainHash]struct{}
	daaAddedBlocksToDelete map[externalapi.DomainHash]struct{}
	daaScoreLRUCache       *lrucache.LRUCache
	daaAddedBlocksLRUCache *lrucache.LRUCache
}

// New instantiates a new DAABlocksStore
func New(daaScoreCacheSize int, daaAddedBlocksCacheSize int, preallocate bool) model.DAABlocksStore {
	return &daaBlocksStore{
		daaScoreStaging:        make(map[externalapi.DomainHash]uint64),
		daaAddedBlocksStaging:  make(map[externalapi.DomainHash][]*externalapi.DomainHash),
		daaScoreLRUCache:       lrucache.New(daaScoreCacheSize, preallocate),
		daaAddedBlocksLRUCache: lrucache.New(daaAddedBlocksCacheSize, preallocate),
	}
}

func (daas *daaBlocksStore) StageDAAScore(blockHash *externalapi.DomainHash, daaScore uint64) {
	daas.daaScoreStaging[*blockHash] = daaScore
}

func (daas *daaBlocksStore) StageBlockDAAAddedBlocks(blockHash *externalapi.DomainHash,
	addedBlocks []*externalapi.DomainHash) {
	daas.daaAddedBlocksStaging[*blockHash] = externalapi.CloneHashes(addedBlocks)
}

func (daas *daaBlocksStore) IsAnythingStaged() bool {
	return len(daas.daaScoreStaging) != 0 ||
		len(daas.daaAddedBlocksStaging) != 0 ||
		len(daas.daaScoreToDelete) != 0 ||
		len(daas.daaAddedBlocksToDelete) != 0
}

func (daas *daaBlocksStore) Discard() {
	daas.daaScoreStaging = make(map[externalapi.DomainHash]uint64)
	daas.daaAddedBlocksStaging = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
	daas.daaScoreToDelete = make(map[externalapi.DomainHash]struct{})
	daas.daaAddedBlocksToDelete = make(map[externalapi.DomainHash]struct{})
}

func (daas *daaBlocksStore) Commit(dbTx model.DBTransaction) error {
	for hash, daaScore := range daas.daaScoreStaging {
		daaScoreBytes := binaryserialization.SerializeUint64(daaScore)
		err := dbTx.Put(daas.daaScoreHashAsKey(&hash), daaScoreBytes)
		if err != nil {
			return err
		}
		daas.daaScoreLRUCache.Add(&hash, daaScore)
	}

	for hash, addedBlocks := range daas.daaAddedBlocksStaging {
		addedBlocksBytes := binaryserialization.SerializeHashes(addedBlocks)
		err := dbTx.Put(daas.daaAddedBlocksHashAsKey(&hash), addedBlocksBytes)
		if err != nil {
			return err
		}
		daas.daaAddedBlocksLRUCache.Add(&hash, addedBlocks)
	}

	for hash := range daas.daaScoreToDelete {
		err := dbTx.Delete(daas.daaScoreHashAsKey(&hash))
		if err != nil {
			return err
		}
		daas.daaScoreLRUCache.Remove(&hash)
	}

	for hash := range daas.daaAddedBlocksToDelete {
		err := dbTx.Delete(daas.daaAddedBlocksHashAsKey(&hash))
		if err != nil {
			return err
		}
		daas.daaAddedBlocksLRUCache.Remove(&hash)
	}

	daas.Discard()
	return nil
}

func (daas *daaBlocksStore) DAAScore(dbContext model.DBReader, blockHash *externalapi.DomainHash) (uint64, error) {
	if daaScore, ok := daas.daaScoreStaging[*blockHash]; ok {
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

func (daas *daaBlocksStore) DAAAddedBlocks(dbContext model.DBReader, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	if addedBlocks, ok := daas.daaAddedBlocksStaging[*blockHash]; ok {
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
	return daaScoreBucket.Key(hash.ByteSlice())
}

func (daas *daaBlocksStore) daaAddedBlocksHashAsKey(hash *externalapi.DomainHash) model.DBKey {
	return daaAddedBlocksBucket.Key(hash.ByteSlice())
}

func (daas *daaBlocksStore) Delete(blockHash *externalapi.DomainHash) {
	if _, ok := daas.daaScoreStaging[*blockHash]; ok {
		delete(daas.daaScoreStaging, *blockHash)
	} else {
		daas.daaAddedBlocksToDelete[*blockHash] = struct{}{}
	}

	if _, ok := daas.daaAddedBlocksStaging[*blockHash]; ok {
		delete(daas.daaAddedBlocksStaging, *blockHash)
	} else {
		daas.daaAddedBlocksToDelete[*blockHash] = struct{}{}
	}
}
