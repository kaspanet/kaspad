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
	daaScoreLRUCache       *lrucache.LRUCache
	daaAddedBlocksLRUCache *lrucache.LRUCache
}

// New instantiates a new BlockRelationStore
func New(cacheSize int, preallocate bool) model.DAABlocksStore {
	return &daaBlocksStore{
		daaScoreStaging:        make(map[externalapi.DomainHash]uint64),
		daaAddedBlocksStaging:  make(map[externalapi.DomainHash][]*externalapi.DomainHash),
		daaScoreLRUCache:       lrucache.New(cacheSize, preallocate),
		daaAddedBlocksLRUCache: lrucache.New(cacheSize, preallocate),
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
	return len(daas.daaScoreStaging) != 0 || len(daas.daaAddedBlocksStaging) != 0
}

func (daas *daaBlocksStore) Discard() {
	daas.daaScoreStaging = make(map[externalapi.DomainHash]uint64)
	daas.daaAddedBlocksStaging = make(map[externalapi.DomainHash][]*externalapi.DomainHash)
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
