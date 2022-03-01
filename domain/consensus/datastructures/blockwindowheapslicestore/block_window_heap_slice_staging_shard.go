package blockwindowheapslicestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type shardKey struct {
	hash       externalapi.DomainHash
	windowSize int
}

type blockWindowHeapSliceStagingShard struct {
	store *blockWindowHeapSliceStore
	toAdd map[shardKey][]*externalapi.BlockGHOSTDAGDataHashPair
}

func (bss *blockWindowHeapSliceStore) stagingShard(stagingArea *model.StagingArea) *blockWindowHeapSliceStagingShard {
	return stagingArea.GetOrCreateShard(bss.shardID, func() model.StagingShard {
		return &blockWindowHeapSliceStagingShard{
			store: bss,
			toAdd: make(map[shardKey][]*externalapi.BlockGHOSTDAGDataHashPair),
		}
	}).(*blockWindowHeapSliceStagingShard)
}

func (bsss *blockWindowHeapSliceStagingShard) Commit(_ model.DBTransaction) error {
	for key, heapSlice := range bsss.toAdd {
		bsss.store.cache.Add(&key.hash, key.windowSize, heapSlice)
	}

	return nil
}

func (bsss *blockWindowHeapSliceStagingShard) isStaged() bool {
	return len(bsss.toAdd) != 0
}

func newShardKey(hash *externalapi.DomainHash, windowSize int) shardKey {
	return shardKey{
		hash:       *hash,
		windowSize: windowSize,
	}
}
